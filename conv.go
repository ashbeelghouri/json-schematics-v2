package schematics

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// toFloat converts any numeric value (including json.Number) to float64.
func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int8:
		return float64(n), true
	case int16:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint8:
		return float64(n), true
	case uint16:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}

func asString(v any) (string, bool) {
	s, ok := v.(string)
	return s, ok
}

// asSlice normalizes any slice value into []any.
func asSlice(v any) ([]any, bool) {
	switch s := v.(type) {
	case []any:
		return s, true
	case []string:
		out := make([]any, len(s))
		for i := range s {
			out[i] = s[i]
		}
		return out, true
	case []int:
		out := make([]any, len(s))
		for i := range s {
			out[i] = s[i]
		}
		return out, true
	case []float64:
		out := make([]any, len(s))
		for i := range s {
			out[i] = s[i]
		}
		return out, true
	default:
		return nil, false
	}
}

func toString(v any) string {
	switch s := v.(type) {
	case string:
		return s
	case fmt.Stringer:
		return s.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

// dateLayouts is the set of layouts parseDate understands.
var dateLayouts = []string{
	"2006-01-02",
	"2006-01-02 15:04:05",
	time.RFC3339,
	time.RFC3339Nano,
	time.RFC1123,
	time.RFC1123Z,
	time.RFC822,
	time.RFC822Z,
	time.RFC850,
	time.ANSIC,
	time.UnixDate,
	time.RubyDate,
	time.Kitchen,
	time.Stamp,
	time.StampMilli,
}

// parseDate parses a value into a time.Time. It accepts time.Time directly, a
// Unix timestamp (number), or a string in any layout in dateLayouts.
func parseDate(v any) (time.Time, bool) {
	switch d := v.(type) {
	case time.Time:
		return d, true
	case string:
		for _, layout := range dateLayouts {
			if t, err := time.Parse(layout, d); err == nil {
				return t, true
			}
		}
		return time.Time{}, false
	default:
		if f, ok := toFloat(v); ok {
			return time.Unix(int64(f), 0).UTC(), true
		}
		return time.Time{}, false
	}
}

func itoa(i int) string { return strconv.Itoa(i) }
