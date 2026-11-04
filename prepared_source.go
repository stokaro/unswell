package unswell

import (
	"crypto/sha256"
	"fmt"
	"unicode/utf8"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/nlp"
)

// PreparedSourceSegment is an exact, owned slice of the original UTF-8 source.
// Its span uses the original source's half-open byte offsets. Text may contain
// markup or escapes that differ from the extracted text used for measurements.
type PreparedSourceSegment struct {
	Span document.Span
	Text string
}

// OriginalPreparedUnit contains the original source segments for one prepared
// target and its enclosing prose. Separate segments must not be joined across
// gaps: those gaps can contain protected code or other excluded source text.
type OriginalPreparedUnit struct {
	Target  []PreparedSourceSegment
	Context []PreparedSourceSegment
}

// OriginalUnit returns exact source segments for one prepared target. It checks
// the source name and SHA-256, the unit contract, UTF-8 boundaries, and the
// target/context relationship before exposing any text. It does not re-extract
// prose or reconstruct the normalized text used by the NLP provider.
func (s PreparedFeatureSource) OriginalUnit(source document.Source, index int) (OriginalPreparedUnit, error) {
	if index < 0 || index >= len(s.Units) || s.TargetCount != len(s.Units) {
		return OriginalPreparedUnit{}, fmt.Errorf("invalid prepared unit index or target count")
	}
	if err := s.validateOriginalSource(source); err != nil {
		return OriginalPreparedUnit{}, err
	}
	binding := s.Units[index].Binding
	if !validOriginalBinding(source.Bytes, binding) {
		return OriginalPreparedUnit{}, fmt.Errorf("invalid prepared target or context spans")
	}
	return OriginalPreparedUnit{
		Target:  originalSourceSegments(source.Bytes, binding.Segments),
		Context: originalSourceSegments(source.Bytes, binding.ContextSpans),
	}, nil
}

func (s PreparedFeatureSource) validateOriginalSource(source document.Source) error {
	if source.Name != s.Path || fmt.Sprintf("%x", sha256.Sum256(source.Bytes)) != s.SourceHash {
		return fmt.Errorf("prepared source identity does not match")
	}
	if !utf8.Valid(source.Bytes) {
		return fmt.Errorf("prepared source is not valid UTF-8")
	}
	return nil
}

func validOriginalBinding(source []byte, binding nlp.UnitBinding) bool {
	return binding.Contract == nlp.UnitContract &&
		validOriginalSpans(source, binding.Segments) &&
		validOriginalSpans(source, binding.ContextSpans) &&
		originalSpanSubset(binding.Segments, binding.ContextSpans)
}

func validOriginalSpans(source []byte, spans []document.Span) bool {
	if len(spans) == 0 || len(spans) > 4096 {
		return false
	}
	last := 0
	for _, span := range spans {
		if span.Start < last || !span.Valid(len(source)) ||
			!utf8.RuneStart(source[span.Start]) ||
			(span.End < len(source) && !utf8.RuneStart(source[span.End])) {
			return false
		}
		last = span.End
	}
	return true
}

func originalSpanSubset(inner, outer []document.Span) bool {
	index := 0
	for _, span := range inner {
		for index < len(outer) && outer[index].End <= span.Start {
			index++
		}
		if index == len(outer) || span.Start < outer[index].Start || span.End > outer[index].End {
			return false
		}
	}
	return true
}

func originalSourceSegments(source []byte, spans []document.Span) []PreparedSourceSegment {
	segments := make([]PreparedSourceSegment, len(spans))
	for i, span := range spans {
		segments[i] = PreparedSourceSegment{Span: span, Text: string(source[span.Start:span.End])}
	}
	return segments
}
