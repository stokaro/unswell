package feature

import (
	"compress/zlib"
	"context"
	"fmt"
)

type compressionSizer struct {
	ctx    context.Context
	writer *zlib.Writer
	bytes  int
}

func newCompressionSizer(ctx context.Context, level int) (*compressionSizer, error) {
	s := &compressionSizer{ctx: ctx}
	writer, err := zlib.NewWriterLevel(s, level)
	s.writer = writer
	return s, err
}

func (s *compressionSizer) Write(data []byte) (int, error) {
	if err := s.ctx.Err(); err != nil {
		return 0, err
	}
	s.bytes += len(data)
	return len(data), nil
}

func (s *compressionSizer) size(parts ...string) (int, error) {
	s.bytes = 0
	s.writer.Reset(s)
	for _, part := range parts {
		for len(part) > 0 {
			if err := s.ctx.Err(); err != nil {
				return 0, err
			}
			n := min(len(part), 4096)
			if _, err := s.writer.Write([]byte(part[:n])); err != nil {
				return 0, err
			}
			part = part[n:]
		}
	}
	if err := s.writer.Close(); err != nil {
		return 0, err
	}
	return s.bytes, s.ctx.Err()
}

func (m *Compression) counts(ctx context.Context, text string) (CompressionCounts, error) {
	if len(m.prefix)+1+2*len(text) > m.identity.Options.MaxInputBytes {
		return CompressionCounts{}, fmt.Errorf("compression target exceeds input-byte budget")
	}
	sizer, err := newCompressionSizer(ctx, m.identity.Options.Level)
	if err != nil {
		return CompressionCounts{}, err
	}
	seeded, err := sizer.size(m.prefix, text)
	if err != nil {
		return CompressionCounts{}, err
	}
	control, err := sizer.size("\n", text)
	if err != nil {
		return CompressionCounts{}, err
	}
	return CompressionCounts{TargetBytes: len(text), SeedBytes: m.seedSize, SeedTargetBytes: seeded,
		ControlBytes: m.baseSize, ControlTargetBytes: control}, nil
}

func compressionValues(c CompressionCounts) []Value {
	seeded, control := c.SeedTargetBytes-c.SeedBytes, c.ControlTargetBytes-c.ControlBytes
	increment := float64(seeded) / float64(c.TargetBytes)
	gain := float64(control-seeded) / float64(c.TargetBytes)
	values := []Value{
		{ID: "compression.incremental-bytes", Version: "1", Unit: "compressed-bytes/input-byte", Number: &increment},
		{ID: "compression.reference-gain", Version: "1", Unit: "compressed-bytes/input-byte", Number: &gain},
	}
	if seeded <= 0 || control <= 0 {
		for i := range values {
			values[i].Number, values[i].Reason = nil, "compression_boundary"
		}
	}
	return values
}
