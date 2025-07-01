package sheetkv

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Record struct {
	Key    int                    // 行番号 (2から始まる、1行目はカラム定義)
	Values map[string]interface{} // カラム名と値のマップ
}

// GetAsString returns the value as string or defaultValue if not found
func (r *Record) GetAsString(col string, defaultValue string) string {
	v, ok := r.Values[col]
	if !ok {
		return defaultValue
	}

	switch val := v.(type) {
	case string:
		return val
	case int, int64, float64:
		return fmt.Sprintf("%v", val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	case []string:
		return strings.Join(val, ",")
	default:
		return fmt.Sprintf("%v", val)
	}
}

// GetAsInt64 returns the value as int64 or defaultValue if not found
func (r *Record) GetAsInt64(col string, defaultValue int64) int64 {
	v, ok := r.Values[col]
	if !ok {
		return defaultValue
	}

	switch val := v.(type) {
	case int64:
		return val
	case int:
		return int64(val)
	case float64:
		return int64(val)
	case string:
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			return i
		}
	}
	return defaultValue
}

// GetAsFloat64 returns the value as float64 or defaultValue if not found
func (r *Record) GetAsFloat64(col string, defaultValue float64) float64 {
	v, ok := r.Values[col]
	if !ok {
		return defaultValue
	}

	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return defaultValue
}

// GetAsStrings returns the value as []string or defaultValue if not found
func (r *Record) GetAsStrings(col string, defaultValue []string) []string {
	v, ok := r.Values[col]
	if !ok {
		return defaultValue
	}

	switch val := v.(type) {
	case []string:
		return val
	case string:
		if val == "" {
			return []string{}
		}
		return strings.Split(val, ",")
	case []interface{}:
		result := make([]string, len(val))
		for i, item := range val {
			result[i] = fmt.Sprintf("%v", item)
		}
		return result
	}
	return defaultValue
}

// GetAsBool returns the value as bool or defaultValue if not found
func (r *Record) GetAsBool(col string, defaultValue bool) bool {
	v, ok := r.Values[col]
	if !ok {
		return defaultValue
	}

	switch val := v.(type) {
	case bool:
		return val
	case string:
		return val == "true" || val == "1"
	case int, int64:
		return val != 0
	case float64:
		return val != 0
	}
	return defaultValue
}

// GetAsTime returns the value as time.Time or defaultValue if not found
func (r *Record) GetAsTime(col string, defaultValue time.Time) time.Time {
	v, ok := r.Values[col]
	if !ok {
		return defaultValue
	}

	switch val := v.(type) {
	case time.Time:
		return val
	case string:
		// Try various formats
		formats := []string{
			time.RFC3339,
			"2006-01-02 15:04:05",
			"2006-01-02",
		}
		for _, format := range formats {
			if t, err := time.Parse(format, val); err == nil {
				return t
			}
		}
	}
	return defaultValue
}

// SetString sets a string value
func (r *Record) SetString(col string, value string) {
	if r.Values == nil {
		r.Values = make(map[string]interface{})
	}
	r.Values[col] = value
}

// SetInt64 sets an int64 value
func (r *Record) SetInt64(col string, value int64) {
	if r.Values == nil {
		r.Values = make(map[string]interface{})
	}
	r.Values[col] = value
}

// SetFloat64 sets a float64 value
func (r *Record) SetFloat64(col string, value float64) {
	if r.Values == nil {
		r.Values = make(map[string]interface{})
	}
	r.Values[col] = value
}

// SetStrings sets a []string value (stored as comma-separated string)
func (r *Record) SetStrings(col string, value []string) {
	if r.Values == nil {
		r.Values = make(map[string]interface{})
	}
	r.Values[col] = strings.Join(value, ",")
}

// SetBool sets a bool value
func (r *Record) SetBool(col string, value bool) {
	if r.Values == nil {
		r.Values = make(map[string]interface{})
	}
	r.Values[col] = value
}

// SetTime sets a time.Time value (stored as ISO 8601 string)
func (r *Record) SetTime(col string, value time.Time) {
	if r.Values == nil {
		r.Values = make(map[string]interface{})
	}
	r.Values[col] = value.Format(time.RFC3339)
}
