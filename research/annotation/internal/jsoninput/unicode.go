package jsoninput

import (
	"fmt"
	"strconv"
)

// Encoding/json replaces lone surrogate escapes. Reject them before that lossy
// conversion so an annotation packet cannot silently change source prose.
func validSurrogates(data []byte) error {
	for i := 0; i < len(data); i++ {
		if data[i] != '\\' || i+1 >= len(data) {
			continue
		}
		i++
		if data[i] != 'u' {
			continue
		}
		next, err := surrogateEnd(data, i)
		if err != nil {
			return err
		}
		i = next
	}
	return nil
}

func surrogateEnd(data []byte, index int) (int, error) {
	value, err := escapeValue(data, index)
	if err != nil {
		return 0, err
	}
	end := index + 4
	if value >= 0xdc00 && value <= 0xdfff {
		return 0, fmt.Errorf("unpaired low surrogate in annotation JSON")
	}
	if value < 0xd800 || value > 0xdbff {
		return end, nil
	}
	return lowSurrogateEnd(data, end)
}

func lowSurrogateEnd(data []byte, end int) (int, error) {
	if end+2 >= len(data) || data[end+1] != '\\' || data[end+2] != 'u' {
		return 0, fmt.Errorf("unpaired high surrogate in annotation JSON")
	}
	low, err := escapeValue(data, end+2)
	if err != nil || low < 0xdc00 || low > 0xdfff {
		return 0, fmt.Errorf("invalid surrogate pair in annotation JSON")
	}
	return end + 6, nil
}

func escapeValue(data []byte, index int) (uint64, error) {
	if index+5 > len(data) {
		return 0, fmt.Errorf("truncated Unicode escape")
	}
	return strconv.ParseUint(string(data[index+1:index+5]), 16, 16)
}
