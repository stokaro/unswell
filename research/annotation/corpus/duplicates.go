package corpus

import (
	"context"
	"encoding/binary"
	"fmt"
	"hash"
	"hash/fnv"
	"slices"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// DuplicatesVersion identifies the near-duplicate detector and its report.
const DuplicatesVersion = "unswell-near-duplicates-v1"

// Detector parameters are part of the version; changing one needs a new version.
const (
	duplicateShingleWords = 8
	duplicateHashes       = 128
	duplicateBands        = 32
	duplicateRows         = duplicateHashes / duplicateBands
	MaxDuplicatePairs     = 100000
)

// DefaultDuplicateThreshold is the Jaccard similarity at which two sources
// count as near duplicates unless a manifest records another value.
const DefaultDuplicateThreshold = 0.5

// DuplicatePair records one candidate pair the detector compared exactly.
type DuplicatePair struct {
	A                string  `json:"a"`
	B                string  `json:"b"`
	EstimatedJaccard float64 `json:"estimated_jaccard"`
	ExactJaccard     float64 `json:"exact_jaccard"`
}

// DuplicateCluster is a connected set of sources above the threshold and the
// related-version key a curator adds to each member.
type DuplicateCluster struct {
	Key     string   `json:"key"`
	Sources []string `json:"sources"`
}

// DuplicateReport is the detector's output for one manifest. It is evidence
// for curation, never an automatic grouping: the planner connects sources
// only through keys the manifest declares.
type DuplicateReport struct {
	Version        string             `json:"version"`
	ManifestSHA256 string             `json:"manifest_sha256"`
	ShingleWords   int                `json:"shingle_words"`
	Hashes         int                `json:"hashes"`
	Bands          int                `json:"bands"`
	Threshold      float64            `json:"threshold"`
	Sources        int                `json:"sources"`
	TooShort       []string           `json:"too_short"`
	Pairs          []DuplicatePair    `json:"pairs"`
	Clusters       []DuplicateCluster `json:"clusters"`
}

// DetectDuplicates shingles the normalized words of every source, finds
// candidate pairs with MinHash bands, measures each candidate exactly, and
// clusters pairs at or above the threshold. Sources are caller-owned bytes
// that must match the manifest; the detector never reads paths.
func DetectDuplicates(ctx context.Context, manifest Manifest, files map[string][]byte, threshold float64) (DuplicateReport, error) {
	if threshold <= 0 || threshold > 1 {
		return DuplicateReport{}, fmt.Errorf("duplicate threshold must be within (0, 1]")
	}
	canonical, err := canonicalManifest(manifest)
	if err != nil {
		return DuplicateReport{}, err
	}
	if err := canonical.validate(ctx); err != nil {
		return DuplicateReport{}, err
	}
	hash, err := digest(canonical)
	if err != nil {
		return DuplicateReport{}, err
	}
	report := DuplicateReport{Version: DuplicatesVersion, ManifestSHA256: hash, ShingleWords: duplicateShingleWords,
		Hashes: duplicateHashes, Bands: duplicateBands, Threshold: threshold, Sources: len(canonical.Sources),
		TooShort: []string{}, Pairs: []DuplicatePair{}, Clusters: []DuplicateCluster{}}
	signatures, shingles, err := signSources(ctx, canonical, files, &report)
	if err != nil {
		return DuplicateReport{}, err
	}
	candidates, err := bandCandidates(ctx, signatures)
	if err != nil {
		return DuplicateReport{}, err
	}
	for _, pair := range candidates {
		if err := ctx.Err(); err != nil {
			return DuplicateReport{}, err
		}
		estimate := estimateJaccard(signatures[pair[0]].values, signatures[pair[1]].values)
		exact := exactJaccard(shingles[pair[0]], shingles[pair[1]])
		report.Pairs = append(report.Pairs, DuplicatePair{A: signatures[pair[0]].id, B: signatures[pair[1]].id,
			EstimatedJaccard: estimate, ExactJaccard: exact})
	}
	report.Clusters = duplicateClusters(report.Pairs, threshold)
	return report, ctx.Err()
}

// ApplyDuplicates returns the manifest with each cluster key added to the
// related-version keys of its members, so the planner connects them.
func ApplyDuplicates(ctx context.Context, manifest Manifest, report DuplicateReport) (Manifest, error) {
	canonical, err := canonicalManifest(manifest)
	if err != nil {
		return Manifest{}, err
	}
	hash, err := digest(canonical)
	if err != nil {
		return Manifest{}, err
	}
	if report.Version != DuplicatesVersion || report.ManifestSHA256 != hash {
		return Manifest{}, fmt.Errorf("duplicate report does not belong to this manifest")
	}
	keys := make(map[string][]string)
	for _, cluster := range report.Clusters {
		for _, id := range cluster.Sources {
			keys[id] = append(keys[id], cluster.Key)
		}
	}
	for i := range canonical.Sources {
		if err := ctx.Err(); err != nil {
			return Manifest{}, err
		}
		addRelatedKeys(&canonical.Sources[i], keys[canonical.Sources[i].ID])
	}
	if err := canonical.validate(ctx); err != nil {
		return Manifest{}, err
	}
	return canonical, nil
}

func addRelatedKeys(source *Source, keys []string) {
	for _, key := range keys {
		if !slices.Contains(source.Related, key) {
			source.Related = append(source.Related, key)
		}
	}
	slices.Sort(source.Related)
}

// hashWrite feeds bytes to an FNV hasher, whose Write never fails.
func hashWrite(hasher hash.Hash64, data []byte) {
	_, _ = hasher.Write(data)
}

type signature struct {
	id     string
	values [duplicateHashes]uint64
}

func signSources(ctx context.Context, manifest Manifest, files map[string][]byte,
	report *DuplicateReport,
) ([]signature, [][]uint64, error) {
	signatures := make([]signature, 0, len(manifest.Sources))
	shingles := make([][]uint64, 0, len(manifest.Sources))
	for _, source := range manifest.Sources {
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}
		data, found := files[source.Path]
		if !found || len(data) != source.Bytes || hashBytes(data) != source.SHA256 {
			return nil, nil, fmt.Errorf("source %s is missing or differs from its declared bytes", source.ID)
		}
		set := shingleSet(data)
		if len(set) == 0 {
			report.TooShort = append(report.TooShort, source.ID)
			continue
		}
		signatures = append(signatures, signature{id: source.ID, values: minHash(set)})
		shingles = append(shingles, set)
	}
	return signatures, shingles, nil
}

// shingleSet returns the sorted distinct hashes of the normalized word
// 8-grams of a source. Normalization lowercases and keeps letters and digits.
func shingleSet(data []byte) []uint64 {
	words := normalizedWords(data)
	if len(words) < duplicateShingleWords {
		return nil
	}
	set := make([]uint64, 0, len(words)-duplicateShingleWords+1)
	for i := 0; i+duplicateShingleWords <= len(words); i++ {
		hasher := fnv.New64a()
		for _, word := range words[i : i+duplicateShingleWords] {
			hashWrite(hasher, []byte(word))
			hashWrite(hasher, []byte{0})
		}
		set = append(set, hasher.Sum64())
	}
	slices.Sort(set)
	return slices.Compact(set)
}

func normalizedWords(data []byte) []string {
	text := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return unicode.ToLower(r)
		}
		return ' '
	}, string(data))
	return strings.Fields(text)
}

// minHash applies duplicateHashes fixed mixing functions to every shingle
// and keeps the minimum of each; the seeds are part of the version.
func minHash(set []uint64) [duplicateHashes]uint64 {
	var values [duplicateHashes]uint64
	for i := range values {
		values[i] = ^uint64(0)
	}
	for _, shingle := range set {
		for i := range values {
			if mixed := mix(shingle ^ duplicateSeed(i)); mixed < values[i] {
				values[i] = mixed
			}
		}
	}
	return values
}

func duplicateSeed(index int) uint64 {
	hasher := fnv.New64a()
	hashWrite(hasher, []byte(DuplicatesVersion))
	hashWrite(hasher, []byte{0})
	hashWrite(hasher, []byte(strconv.Itoa(index)))
	return hasher.Sum64()
}

// mix is the SplitMix64 finalizer; it spreads shingle hashes per seed.
func mix(value uint64) uint64 {
	value ^= value >> 30
	value *= 0xbf58476d1ce4e5b9
	value ^= value >> 27
	value *= 0x94d049bb133111eb
	value ^= value >> 31
	return value
}

// bandCandidates returns index pairs that agree on at least one band of
// duplicateRows consecutive signature values, in ascending order.
func bandCandidates(ctx context.Context, signatures []signature) ([][2]int, error) {
	seen := make(map[[2]int]bool)
	for band := range duplicateBands {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if err := collectBand(signatures, band, seen); err != nil {
			return nil, err
		}
	}
	pairs := make([][2]int, 0, len(seen))
	for pair := range seen {
		pairs = append(pairs, pair)
	}
	sort.Slice(pairs, func(a, b int) bool {
		if pairs[a][0] != pairs[b][0] {
			return pairs[a][0] < pairs[b][0]
		}
		return pairs[a][1] < pairs[b][1]
	})
	return pairs, nil
}

func collectBand(signatures []signature, band int, seen map[[2]int]bool) error {
	buckets := make(map[uint64][]int)
	for i, item := range signatures {
		key := bandKey(item, band)
		for _, other := range buckets[key] {
			seen[[2]int{other, i}] = true
			if len(seen) > MaxDuplicatePairs {
				return fmt.Errorf("candidate pairs exceed %d", MaxDuplicatePairs)
			}
		}
		buckets[key] = append(buckets[key], i)
	}
	return nil
}

func bandKey(item signature, band int) uint64 {
	hasher := fnv.New64a()
	var buffer [8]byte
	for _, value := range item.values[band*duplicateRows : (band+1)*duplicateRows] {
		binary.BigEndian.PutUint64(buffer[:], value)
		hashWrite(hasher, buffer[:])
	}
	return hasher.Sum64()
}

func estimateJaccard(a, b [duplicateHashes]uint64) float64 {
	matches := 0
	for i := range a {
		if a[i] == b[i] {
			matches++
		}
	}
	return float64(matches) / duplicateHashes
}

func exactJaccard(a, b []uint64) float64 {
	shared, i, j := 0, 0, 0
	for i < len(a) && j < len(b) {
		switch {
		case a[i] == b[j]:
			shared, i, j = shared+1, i+1, j+1
		case a[i] < b[j]:
			i++
		default:
			j++
		}
	}
	union := len(a) + len(b) - shared
	if union == 0 {
		return 0
	}
	return float64(shared) / float64(union)
}

func duplicateClusters(pairs []DuplicatePair, threshold float64) []DuplicateCluster {
	parent := make(map[string]string)
	var find func(string) string
	find = func(id string) string {
		if parent[id] == "" || parent[id] == id {
			parent[id] = id
			return id
		}
		parent[id] = find(parent[id])
		return parent[id]
	}
	for _, pair := range pairs {
		if pair.ExactJaccard < threshold {
			continue
		}
		a, b := find(pair.A), find(pair.B)
		if a != b {
			parent[max(a, b)] = min(a, b)
		}
	}
	members := make(map[string][]string)
	for id := range parent {
		root := find(id)
		members[root] = append(members[root], id)
	}
	clusters := make([]DuplicateCluster, 0, len(members))
	for _, ids := range members {
		slices.Sort(ids)
		hasher := fnv.New64a()
		for _, id := range ids {
			hashWrite(hasher, []byte(id))
			hashWrite(hasher, []byte{0})
		}
		clusters = append(clusters, DuplicateCluster{Key: fmt.Sprintf("near-duplicate-v1:%016x", hasher.Sum64()), Sources: ids})
	}
	slices.SortFunc(clusters, func(a, b DuplicateCluster) int { return strings.Compare(a.Key, b.Key) })
	return clusters
}
