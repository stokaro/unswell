package feature

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/stokaro/unswell/nlp"
)

// Validate checks settings independent of the reference and target size.
func (o CompressionOptions) Validate() error {
	if o.Level < 0 || o.Level > 9 || o.MaxInputBytes < 1 || o.MaxInputBytes > 1<<20 {
		return fmt.Errorf("invalid compression level or input-byte budget")
	}
	return nil
}

func compressionPrefix(reference string, options CompressionOptions) (string, error) {
	if err := options.Validate(); err != nil {
		return "", err
	}
	if len(reference) == 0 || len(reference) >= 32<<10 || !utf8.ValidString(reference) || strings.ContainsRune(reference, 0) {
		return "", fmt.Errorf("compression reference requires nonempty UTF-8 without protected boundaries and a prefix within 32KiB")
	}
	prefix := reference + "\n"
	if len(prefix)+1 > options.MaxInputBytes {
		return "", fmt.Errorf("compression baseline exceeds input-byte budget")
	}
	return prefix, nil
}

func compressionTarget(ctx context.Context, unit nlp.PreparedUnit, identity Identity, limits Limits) (string, error) {
	if unit.Binding().Contract != nlp.UnitContract {
		return "", fmt.Errorf("compression requires a prepared target")
	}
	if err := matchUnitNLP(unit, identity); err != nil {
		return "", err
	}
	block := unit.Block()
	if err := validateInputs(ctx, block, identity, limits); err != nil {
		return "", err
	}
	if len(block.Text) == 0 || len(block.Text) > 64<<10 || strings.ContainsRune(block.Text, 0) {
		return "", fmt.Errorf("compression requires one complete target within 64KiB")
	}
	return block.Text, nil
}
