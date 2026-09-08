package cli

import (
	"bytes"
	"context"
	"crypto/sha1" // #nosec G505 -- Required to verify existing Git SHA-1 object identities.
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const gitListLimit = 16 << 20

type gitEntry struct {
	mode string
	id   string
}

type gitComparison struct {
	root, project, prefix string
	base, head            string
	before, after, index  map[string]gitEntry
}

type limitedGitOutput struct {
	bytes.Buffer
	limit int
}

func (b *limitedGitOutput) Write(data []byte) (int, error) {
	if len(data) > b.limit-b.Len() {
		return 0, fmt.Errorf("git output exceeds %d bytes", b.limit)
	}
	return b.Buffer.Write(data)
}

func gitOutput(ctx context.Context, root string, limit int, args ...string) ([]byte, error) {
	argv := append([]string{"--no-pager", "-c", "core.fsmonitor=false", "-C", root}, args...)
	// #nosec G204 -- Read-only Git operations use argv and validated object IDs, never a shell or source executable.
	command := exec.CommandContext(ctx, "git", argv...)
	command.WaitDelay = time.Second
	command.Env = gitEnvironment()
	output, diagnostic := &limitedGitOutput{limit: limit}, &limitedGitOutput{limit: 4096}
	command.Stdout, command.Stderr = output, diagnostic
	if err := command.Run(); err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("git %s: %w: %s", args[0], err, strings.TrimSpace(diagnostic.String()))
	}
	return output.Bytes(), nil
}

func gitEnvironment() []string {
	var env []string
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(strings.ToUpper(entry), "GIT_") {
			env = append(env, entry)
		}
	}
	return append(env, "GIT_TERMINAL_PROMPT=0", "GIT_OPTIONAL_LOCKS=0", "GIT_NO_REPLACE_OBJECTS=1", "GIT_NO_LAZY_FETCH=1")
}

func gitRevision(ctx context.Context, root, ref string) (string, error) {
	output, err := gitOutput(ctx, root, 1024, "rev-parse", "--verify", "--end-of-options", ref+"^{commit}")
	if err != nil {
		return "", err
	}
	id := strings.TrimSpace(string(output))
	if !gitObjectID(id) {
		return "", fmt.Errorf("git returned an invalid commit ID")
	}
	return id, nil
}

func gitObjectID(id string) bool {
	if len(id) != 40 && len(id) != 64 {
		return false
	}
	for _, char := range id {
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') {
			return false
		}
	}
	return true
}

func openGitComparison(ctx context.Context, project, ref string) (*gitComparison, error) {
	output, err := gitOutput(ctx, project, 1<<20, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, err
	}
	root := strings.TrimSuffix(string(output), "\n")
	resolved, err := filepath.EvalSymlinks(project)
	if err != nil {
		return nil, err
	}
	prefix, err := filepath.Rel(root, resolved)
	if err != nil || escapesRoot(prefix) {
		return nil, fmt.Errorf("project root must be inside the Git worktree")
	}
	if prefix == "." {
		prefix = ""
	} else {
		prefix = filepath.ToSlash(prefix) + "/"
	}
	comparison := &gitComparison{root: root, project: project, prefix: prefix}
	if err := comparison.revisions(ctx, ref); err != nil {
		return nil, err
	}
	if err := comparison.entries(ctx); err != nil {
		return nil, err
	}
	return comparison, nil
}

func (g *gitComparison) revisions(ctx context.Context, ref string) error {
	head, err := gitRevision(ctx, g.root, "HEAD")
	if err != nil {
		return err
	}
	reference, err := gitRevision(ctx, g.root, ref)
	if err != nil {
		return err
	}
	output, err := gitOutput(ctx, g.root, 1024, "merge-base", "--all", reference, head)
	if err != nil {
		return err
	}
	base := strings.TrimSpace(string(output))
	if !gitObjectID(base) {
		return fmt.Errorf("changed analysis requires exactly one available merge base")
	}
	g.base, g.head = base, head
	return nil
}

func (g *gitComparison) entries(ctx context.Context) error {
	for _, revision := range []struct {
		id   string
		dest *map[string]gitEntry
	}{{g.base, &g.before}, {g.head, &g.after}} {
		data, err := gitOutput(ctx, g.root, gitListLimit, "ls-tree", "-r", "-z", "--full-tree", revision.id)
		if err != nil {
			return err
		}
		*revision.dest, err = gitEntries(data, false)
		if err != nil {
			return err
		}
	}
	data, err := gitOutput(ctx, g.root, gitListLimit, "ls-files", "--stage", "--full-name", "-z")
	if err != nil {
		return err
	}
	g.index, err = gitEntries(data, true)
	return err
}

func gitEntries(data []byte, index bool) (map[string]gitEntry, error) {
	result := make(map[string]gitEntry)
	for entry := range strings.SplitSeq(strings.TrimSuffix(string(data), "\x00"), "\x00") {
		if entry == "" {
			continue
		}
		metadata, name, ok := strings.Cut(entry, "\t")
		fields := strings.Fields(metadata)
		if !ok || len(fields) != 3 || len(result) >= 100000 {
			return nil, fmt.Errorf("invalid Git entries or more than 100000 paths")
		}
		id := fields[2]
		if index {
			id = fields[1]
			if fields[2] != "0" {
				return nil, fmt.Errorf("unmerged Git index: %s", name)
			}
		}
		if !gitObjectID(id) {
			return nil, fmt.Errorf("invalid Git object ID for %s", name)
		}
		result[name] = gitEntry{mode: fields[0], id: id}
	}
	return result, nil
}

func (g *gitComparison) blob(ctx context.Context, entry gitEntry, limit int) ([]byte, error) {
	if !regularGitEntry(entry) {
		return nil, fmt.Errorf("committed sources must be regular files")
	}
	return gitOutput(ctx, g.root, limit, "cat-file", "blob", entry.id)
}

func regularGitEntry(entry gitEntry) bool {
	return entry.mode == "100644" || entry.mode == "100755"
}

func gitBlobID(data []byte, expected string) string {
	header := fmt.Appendf(nil, "blob %d\x00", len(data))
	if len(expected) == 64 {
		hash := sha256.New()
		_, _ = hash.Write(header)
		_, _ = hash.Write(data)
		return fmt.Sprintf("%x", hash.Sum(nil))
	}
	// #nosec G401 -- SHA-1 matches Git's existing object format, not a security or authentication decision.
	hash := sha1.New()
	_, _ = hash.Write(header)
	_, _ = hash.Write(data)
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func (g *gitComparison) verifyFile(name string, data []byte) error {
	entry := g.after[g.prefix+name]
	if !regularGitEntry(entry) || entry != g.index[g.prefix+name] {
		return fmt.Errorf("selected file or index does not match HEAD: %s", name)
	}
	if gitBlobID(data, entry.id) != entry.id {
		return fmt.Errorf("selected source bytes do not match HEAD: %s", name)
	}
	return nil
}

func (g *gitComparison) verifyHead(ctx context.Context) error {
	head, err := gitRevision(ctx, g.root, "HEAD")
	if err != nil {
		return err
	}
	if head != g.head {
		return fmt.Errorf("HEAD changed during analysis")
	}
	return nil
}

func regularWorktreeFile(root, name string, limit int) ([]byte, error) {
	if !filepath.IsLocal(filepath.FromSlash(name)) {
		return nil, fmt.Errorf("invalid worktree path: %s", name)
	}
	for path := filepath.Join(root, filepath.FromSlash(name)); path != root; path = filepath.Dir(path) {
		info, err := os.Lstat(path)
		if err != nil {
			return nil, err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("changed analysis does not follow symlinks: %s", name)
		}
		if path == filepath.Join(root, filepath.FromSlash(name)) && !info.Mode().IsRegular() {
			return nil, fmt.Errorf("selected source is not a regular file: %s", name)
		}
	}
	handle, err := os.OpenRoot(root)
	if err != nil {
		return nil, err
	}
	data, readErr := readWorktreeFile(handle, name, limit)
	return data, errors.Join(readErr, handle.Close())
}

func readWorktreeFile(root *os.Root, name string, limit int) ([]byte, error) {
	file, err := root.Open(filepath.FromSlash(name))
	if err != nil {
		return nil, err
	}
	info, statErr := file.Stat()
	if statErr != nil {
		return nil, errors.Join(statErr, file.Close())
	}
	if !info.Mode().IsRegular() || info.Size() > int64(limit) {
		return nil, errors.Join(fmt.Errorf("selected source exceeds byte limit or is not regular: %s", name), file.Close())
	}
	data, readErr := io.ReadAll(io.LimitReader(file, int64(limit)+1))
	if len(data) > limit {
		readErr = fmt.Errorf("selected source exceeds byte limit: %s", name)
	}
	return data, errors.Join(readErr, file.Close())
}
