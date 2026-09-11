package llmdet

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"sort"

	"github.com/stokaro/unswell/research/annotation/internal/jsoninput"
)

// TableVersion identifies the binary form of one published n-gram
// dictionary: every context and every retained value of the unigram, bigram,
// and trigram tables, sorted for lookup, with 16-bit tokens and the original
// 16-bit probabilities.
const TableVersion = "unswell-llmdet-table-v1"

// MaxTableBytes bounds one table file.
const MaxTableBytes = 2 << 30

var tableMagic = []byte("ULDT\x01")

// TableHeader is the JSON header of a table: the model, the vocabulary size
// the reference passes for it, the source dictionary, and the size of every
// order.
type TableHeader struct {
	Version   string       `json:"version"`
	Model     string       `json:"model"`
	VocabSize int          `json:"vocab_size"`
	Source    TableSource  `json:"source"`
	Orders    []TableOrder `json:"orders"`
}

// TableSource names the dictionary file a table came from.
type TableSource struct {
	File   string `json:"file"`
	SHA256 string `json:"sha256"`
}

// TableOrder counts the contexts and values of one n-gram order.
type TableOrder struct {
	N        int `json:"n"`
	Contexts int `json:"contexts"`
	Values   int `json:"values"`
}

// Table holds one dictionary in memory and measures token sequences under
// the reference's proxy arithmetic. The vocabulary size enters the residual
// mass only, as in the reference; a token the table never saw matches no
// context, whatever its ID.
type Table struct {
	header TableHeader
	orders [3]tableOrder
}

type tableOrder struct {
	n         int
	contexts  []uint32  // n tokens per context, sorted lexicographically
	offsets   []uint32  // len(contexts)/n + 1 cumulative value counts
	tokens    []uint16  // retained continuations
	bits      []uint16  // IEEE half-precision probabilities
	residuals []float64 // per context: the reference's residual mass, or zero
}

// LoadTable reads a table file and checks its header, sizes, tokens, and
// probabilities before it measures anything.
func LoadTable(ctx context.Context, path string) (*Table, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	file, err := os.Open(path) // #nosec G304 -- The developer explicitly selects the local table.
	if err != nil {
		return nil, err
	}
	table, err := readTableFile(ctx, file)
	return table, errors.Join(err, file.Close())
}

func readTableFile(ctx context.Context, file *os.File) (*Table, error) {
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > MaxTableBytes {
		return nil, fmt.Errorf("table must be a regular file within %d bytes", MaxTableBytes)
	}
	return ReadTable(ctx, file)
}

// ReadTable reads a table from a stream.
func ReadTable(ctx context.Context, reader io.Reader) (*Table, error) {
	header, err := readTableHeader(ctx, reader)
	if err != nil {
		return nil, err
	}
	table := &Table{header: header}
	for i, order := range header.Orders {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		loaded, err := readTableOrder(reader, order, header.VocabSize)
		if err != nil {
			return nil, fmt.Errorf("table order %d: %w", order.N, err)
		}
		table.orders[i] = loaded
	}
	if _, err := reader.Read(make([]byte, 1)); err != io.EOF {
		return nil, fmt.Errorf("table has trailing bytes")
	}
	return table, nil
}

func readTableHeader(ctx context.Context, reader io.Reader) (TableHeader, error) {
	prefix := make([]byte, len(tableMagic)+4)
	if _, err := io.ReadFull(reader, prefix); err != nil || !bytes.Equal(prefix[:len(tableMagic)], tableMagic) {
		return TableHeader{}, fmt.Errorf("table lacks its magic prefix")
	}
	length := binary.LittleEndian.Uint32(prefix[len(tableMagic):])
	if length == 0 || length > 1<<16 {
		return TableHeader{}, fmt.Errorf("table header length out of range")
	}
	encoded := make([]byte, length)
	if _, err := io.ReadFull(reader, encoded); err != nil {
		return TableHeader{}, err
	}
	var header TableHeader
	if err := jsoninput.Decode(ctx, encoded, 1<<16, &header, jsoninput.Limits{Array: 3, Object: 8}); err != nil {
		return TableHeader{}, err
	}
	return header, validateTableHeader(header)
}

func validateTableHeader(header TableHeader) error {
	if header.Version != TableVersion || header.Model == "" || header.VocabSize < 2 || header.VocabSize > 1<<16 ||
		len(header.Orders) != 3 || len(header.Source.SHA256) != 64 {
		return fmt.Errorf("unsupported table version, model, vocabulary, or orders")
	}
	for i, order := range header.Orders {
		if order.N != i+1 || !validTableDimensions(order) {
			return fmt.Errorf("table order %d has unsupported dimensions", order.N)
		}
	}
	return nil
}

func validTableDimensions(order TableOrder) bool {
	return order.Contexts >= 0 && order.Values >= 0 && order.Contexts <= 1<<24 && order.Values <= 1<<28
}

func readTableOrder(reader io.Reader, order TableOrder, vocab int) (tableOrder, error) {
	result := tableOrder{n: order.N}
	var err error
	if result.contexts, err = readUint32s(reader, order.Contexts*order.N); err != nil {
		return tableOrder{}, err
	}
	if result.offsets, err = readUint32s(reader, order.Contexts+1); err != nil {
		return tableOrder{}, err
	}
	if result.tokens, err = readUint16s(reader, order.Values); err != nil {
		return tableOrder{}, err
	}
	if result.bits, err = readUint16s(reader, order.Values); err != nil {
		return tableOrder{}, err
	}
	if err := result.validate(order.Values); err != nil {
		return tableOrder{}, err
	}
	result.residuals = make([]float64, order.Contexts)
	for i := range order.Contexts {
		result.residuals[i] = residualMass(result.bits[result.offsets[i]:result.offsets[i+1]], vocab)
	}
	return result, nil
}

func (o tableOrder) validate(values int) error {
	if err := o.validateOffsets(values); err != nil {
		return err
	}
	if err := o.validateContexts(); err != nil {
		return err
	}
	for i := range o.tokens {
		probability := halfToFloat(o.bits[i])
		if math.IsNaN(probability) || math.IsInf(probability, 0) || probability < 0 || probability > 1 {
			return fmt.Errorf("probability outside the unit interval")
		}
	}
	return nil
}

func (o tableOrder) validateOffsets(values int) error {
	if o.offsets[0] != 0 || int(o.offsets[len(o.offsets)-1]) != values {
		return fmt.Errorf("offsets do not span the values")
	}
	for i := 1; i < len(o.offsets); i++ {
		if o.offsets[i] < o.offsets[i-1] {
			return fmt.Errorf("offsets decrease")
		}
	}
	return nil
}

func (o tableOrder) validateContexts() error {
	for _, token := range o.contexts {
		if token >= 1<<16 {
			return fmt.Errorf("context token outside the 16-bit range")
		}
	}
	for i := o.n; i < len(o.contexts); i += o.n {
		if compareContext(o.contexts[i-o.n:i], o.contexts[i:i+o.n]) >= 0 {
			return fmt.Errorf("contexts are not sorted and distinct")
		}
	}
	return nil
}

// residualMass follows the reference: the retained probabilities are summed
// as 16-bit floats, one at a time, and the remainder is spread over the
// tokens the context does not retain.
func residualMass(bits []uint16, vocab int) float64 {
	sum := uint16(0)
	for _, b := range bits {
		sum = floatToHalf(halfToFloat(sum) + halfToFloat(b))
	}
	remaining := 1 - halfToFloat(sum)
	if remaining <= 0 || len(bits) >= vocab {
		return 0
	}
	return remaining / float64(vocab-len(bits))
}

func readUint32s(reader io.Reader, count int) ([]uint32, error) {
	raw := make([]byte, count*4)
	if _, err := io.ReadFull(reader, raw); err != nil {
		return nil, err
	}
	values := make([]uint32, count)
	for i := range values {
		values[i] = binary.LittleEndian.Uint32(raw[i*4:])
	}
	return values, nil
}

func readUint16s(reader io.Reader, count int) ([]uint16, error) {
	raw := make([]byte, count*2)
	if _, err := io.ReadFull(reader, raw); err != nil {
		return nil, err
	}
	values := make([]uint16, count)
	for i := range values {
		values[i] = binary.LittleEndian.Uint16(raw[i*2:])
	}
	return values, nil
}

func compareContext(a, b []uint32) int {
	for i := range a {
		if a[i] != b[i] {
			if a[i] < b[i] {
				return -1
			}
			return 1
		}
	}
	return 0
}

// Header returns the table's header by value.
func (t *Table) Header() TableHeader {
	if t == nil {
		return TableHeader{}
	}
	header := t.header
	header.Orders = append([]TableOrder(nil), t.header.Orders...)
	return header
}

// Measure applies the reference's proxy arithmetic to a token sequence with
// this table's contexts. It returns the same result form as Proxy.Measure.
func (t *Table) Measure(ctx context.Context, tokens []int) (ProxyResult, error) {
	var result ProxyResult
	if err := ctx.Err(); err != nil {
		return result, err
	}
	if t == nil || len(tokens) > MaxTokens {
		return result, fmt.Errorf("missing table or excessive token count")
	}
	result.Possible = max(0, len(tokens)-3)
	sum := 0.0
	for i := 2; i < len(tokens)-1; i++ {
		if err := ctx.Err(); err != nil {
			return ProxyResult{}, err
		}
		row, order, found := t.context(tokens[i-2 : i+1])
		if !found {
			continue
		}
		sum += observeRow(&result, row, order, tokens[i+1])
	}
	result.ReferenceScore = -sum / float64(result.Matched+1)
	finishProxyResult(&result)
	return result, nil
}

// context looks the longest matching context up, as the reference does, and
// materializes its retained values.
func (t *Table) context(tokens []int) (probabilityRow, int, bool) {
	for size := 3; size > 0; size-- {
		order := t.orders[size-1]
		index, found := order.find(tokens[3-size:])
		if !found {
			continue
		}
		start, end := order.offsets[index], order.offsets[index+1]
		row := probabilityRow{values: make(map[int]float64, end-start), residual: order.residuals[index]}
		for i := start; i < end; i++ {
			row.values[int(order.tokens[i])] = halfToFloat(order.bits[i])
		}
		return row, size, true
	}
	return probabilityRow{}, 0, false
}

func (o tableOrder) find(tokens []int) (int, bool) {
	var key [3]uint32
	for i, token := range tokens {
		if token < 0 || token >= 1<<16 {
			return 0, false
		}
		key[i] = uint32(token)
	}
	count := len(o.contexts) / o.n
	index := sort.Search(count, func(i int) bool { return compareContext(o.contexts[i*o.n:(i+1)*o.n], key[:o.n]) >= 0 })
	return index, index < count && compareContext(o.contexts[index*o.n:(index+1)*o.n], key[:o.n]) == 0
}

// halfToFloat decodes an IEEE 754 half-precision value.
func halfToFloat(bits uint16) float64 {
	sign := float64(1)
	if bits&0x8000 != 0 {
		sign = -1
	}
	exponent := int((bits >> 10) & 0x1f)
	mantissa := float64(bits & 0x3ff)
	switch exponent {
	case 0:
		return sign * math.Ldexp(mantissa, -24)
	case 0x1f:
		if mantissa == 0 {
			return sign * math.Inf(1)
		}
		return math.NaN()
	}
	return sign * math.Ldexp(1+mantissa/1024, exponent-15)
}

// floatToHalf rounds a value to the nearest half-precision value, ties to
// even, as the reference's 16-bit accumulation does.
func floatToHalf(value float64) uint16 {
	if math.IsNaN(value) {
		return 0x7e00
	}
	var sign uint16
	if value < 0 || (value == 0 && math.Signbit(value)) {
		sign, value = 0x8000, -value
	}
	if bits, special := halfSpecial(value); special {
		return sign | bits
	}
	exponent := math.Floor(math.Log2(value))
	if math.Ldexp(1, int(exponent)) > value {
		exponent--
	}
	mantissa := roundHalfEven((value/math.Ldexp(1, int(exponent)) - 1) * 1024)
	if mantissa == 1024 {
		mantissa, exponent = 0, exponent+1
	}
	if exponent+15 >= 0x1f {
		return sign | 0x7c00
	}
	return sign | uint16(exponent+15)<<10 | uint16(mantissa)
}

// halfSpecial handles the magnitudes outside the normal half-precision
// range: infinities and overflow, values that round to zero, and subnormals.
func halfSpecial(value float64) (uint16, bool) {
	switch {
	case math.IsInf(value, 0) || value >= 65520:
		return 0x7c00, true
	case value < math.Ldexp(1, -25):
		return 0, true
	case value < math.Ldexp(1, -14):
		return uint16(roundHalfEven(value * math.Ldexp(1, 24))), true // #nosec G115 -- below 1024 by the range check.
	}
	return 0, false
}

func roundHalfEven(value float64) float64 {
	return math.RoundToEven(value)
}

// TableRow is one context of a table under construction: its tokens, its
// retained continuations, and their probabilities in order.
type TableRow struct {
	Context       []int
	Continuations []int
	Probabilities []float64
}

// EncodeTable writes a table in its binary form from rows of any order. It
// sorts the contexts and rounds each probability to half precision, so a
// table built here reads back through ReadTable.
func EncodeTable(model string, vocab int, source TableSource, rows []TableRow) ([]byte, error) {
	if model == "" || vocab < 2 || vocab > 1<<16 {
		return nil, fmt.Errorf("table requires a model name and a 16-bit vocabulary")
	}
	if err := validateTableRows(rows); err != nil {
		return nil, err
	}
	header := TableHeader{Version: TableVersion, Model: model, VocabSize: vocab, Source: source}
	var payload bytes.Buffer
	for n := 1; n <= 3; n++ {
		selected := rowsOfOrder(rows, n)
		values := 0
		for _, row := range selected {
			values += len(row.Continuations)
		}
		header.Orders = append(header.Orders, TableOrder{N: n, Contexts: len(selected), Values: values})
		writeTableOrder(&payload, selected)
	}
	encoded, err := json.Marshal(header)
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	out.Write(tableMagic)
	_ = binary.Write(&out, binary.LittleEndian, uint32(len(encoded))) // #nosec G115 -- a header is far below 4 GiB.
	out.Write(encoded)
	out.Write(payload.Bytes())
	return out.Bytes(), nil
}

func validateTableRows(rows []TableRow) error {
	if len(rows) > 1<<24 {
		return fmt.Errorf("table exceeds its context bound")
	}
	for _, row := range rows {
		if len(row.Context) < 1 || len(row.Context) > 3 {
			return fmt.Errorf("table contexts hold one to three tokens")
		}
		if len(row.Continuations) != len(row.Probabilities) || len(row.Continuations) > 1<<16 {
			return fmt.Errorf("continuations and probabilities differ in length or exceed the bound")
		}
		for _, token := range append(append([]int{}, row.Context...), row.Continuations...) {
			if token < 0 || token >= 1<<16 {
				return fmt.Errorf("token outside the 16-bit range")
			}
		}
	}
	return nil
}

// rowsOfOrder returns the rows of one context length, sorted by context.
func rowsOfOrder(rows []TableRow, n int) []TableRow {
	selected := make([]TableRow, 0)
	for _, row := range rows {
		if len(row.Context) == n {
			selected = append(selected, row)
		}
	}
	sort.SliceStable(selected, func(i, j int) bool {
		return compareInts(selected[i].Context, selected[j].Context) < 0
	})
	return selected
}

// writeTableOrder writes one order; EncodeTable bounded every token and
// count before calling it.
func writeTableOrder(payload *bytes.Buffer, rows []TableRow) {
	for _, row := range rows {
		for _, token := range row.Context {
			_ = binary.Write(payload, binary.LittleEndian, uint32(token)) // #nosec G115 -- bounded by EncodeTable.
		}
	}
	offset := uint32(0)
	_ = binary.Write(payload, binary.LittleEndian, offset)
	for _, row := range rows {
		offset += uint32(len(row.Continuations)) // #nosec G115 -- bounded by EncodeTable.
		_ = binary.Write(payload, binary.LittleEndian, offset)
	}
	for _, row := range rows {
		for _, token := range row.Continuations {
			_ = binary.Write(payload, binary.LittleEndian, uint16(token)) // #nosec G115 -- bounded by EncodeTable.
		}
	}
	for _, row := range rows {
		for _, probability := range row.Probabilities {
			_ = binary.Write(payload, binary.LittleEndian, floatToHalf(probability))
		}
	}
}

func compareInts(a, b []int) int {
	for i := range a {
		if a[i] != b[i] {
			if a[i] < b[i] {
				return -1
			}
			return 1
		}
	}
	return 0
}
