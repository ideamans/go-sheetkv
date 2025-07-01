package sheetkv

import (
	"fmt"
)

// Condition represents a single query condition
type Condition struct {
	Column   string      // カラム名
	Operator string      // 演算子: ==, !=, >, >=, <, <=, in, between
	Value    interface{} // 比較値（inの場合は[]interface{}, betweenの場合は[2]interface{}）
}

// Query represents a query with multiple conditions
type Query struct {
	Conditions []Condition // AND条件として評価
	Limit      int
	Offset     int
}

// evalCondition evaluates a single condition against a record
func evalCondition(record *Record, condition Condition) bool {
	value, exists := record.Values[condition.Column]
	if !exists {
		// カラムが存在しない場合、nullとして扱う
		value = nil
	}

	switch condition.Operator {
	case "==":
		return compareEqual(value, condition.Value)
	case "!=":
		return !compareEqual(value, condition.Value)
	case ">":
		return compareGreater(value, condition.Value)
	case ">=":
		return compareGreaterEqual(value, condition.Value)
	case "<":
		return compareLess(value, condition.Value)
	case "<=":
		return compareLessEqual(value, condition.Value)
	case "in":
		return compareIn(value, condition.Value)
	case "between":
		return compareBetween(value, condition.Value)
	default:
		return false
	}
}

// MatchesQuery checks if a record matches all conditions in the query
func (r *Record) MatchesQuery(query Query) bool {
	// 全ての条件をANDで評価
	for _, condition := range query.Conditions {
		if !evalCondition(r, condition) {
			return false
		}
	}
	return true
}

// compareEqual compares two values for equality
func compareEqual(a, b interface{}) bool {
	// 両方がnilの場合
	if a == nil && b == nil {
		return true
	}
	// 片方だけがnilの場合
	if a == nil || b == nil {
		return false
	}

	// 数値の比較は型変換を考慮
	if isNumeric(a) && isNumeric(b) {
		return toFloat64(a) == toFloat64(b)
	}

	// その他は通常の比較
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

// compareGreater compares if a > b
func compareGreater(a, b interface{}) bool {
	if !isNumeric(a) || !isNumeric(b) {
		return false
	}
	return toFloat64(a) > toFloat64(b)
}

// compareGreaterEqual compares if a >= b
func compareGreaterEqual(a, b interface{}) bool {
	if !isNumeric(a) || !isNumeric(b) {
		return false
	}
	return toFloat64(a) >= toFloat64(b)
}

// compareLess compares if a < b
func compareLess(a, b interface{}) bool {
	if !isNumeric(a) || !isNumeric(b) {
		return false
	}
	return toFloat64(a) < toFloat64(b)
}

// compareLessEqual compares if a <= b
func compareLessEqual(a, b interface{}) bool {
	if !isNumeric(a) || !isNumeric(b) {
		return false
	}
	return toFloat64(a) <= toFloat64(b)
}

// compareIn checks if a is in the list b
func compareIn(a, b interface{}) bool {
	// bは[]interface{}である必要がある
	list, ok := b.([]interface{})
	if !ok {
		return false
	}

	for _, item := range list {
		if compareEqual(a, item) {
			return true
		}
	}
	return false
}

// compareBetween checks if a is between b[0] and b[1]
func compareBetween(a, b interface{}) bool {
	// bは[2]interface{}である必要がある
	var min, max interface{}

	switch v := b.(type) {
	case [2]interface{}:
		min, max = v[0], v[1]
	case []interface{}:
		if len(v) != 2 {
			return false
		}
		min, max = v[0], v[1]
	default:
		return false
	}

	if !isNumeric(a) || !isNumeric(min) || !isNumeric(max) {
		return false
	}

	aVal := toFloat64(a)
	minVal := toFloat64(min)
	maxVal := toFloat64(max)

	return aVal >= minVal && aVal <= maxVal
}

// isNumeric checks if a value is numeric
func isNumeric(v interface{}) bool {
	switch v.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return true
	default:
		return false
	}
}

// toFloat64 converts a numeric value to float64
func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case int:
		return float64(val)
	case int8:
		return float64(val)
	case int16:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	case uint:
		return float64(val)
	case uint8:
		return float64(val)
	case uint16:
		return float64(val)
	case uint32:
		return float64(val)
	case uint64:
		return float64(val)
	case float32:
		return float64(val)
	case float64:
		return val
	default:
		return 0
	}
}

// ApplyQuery filters records based on query conditions
func ApplyQuery(records []*Record, query Query) []*Record {
	var results []*Record

	// フィルタリング
	for _, record := range records {
		if record.MatchesQuery(query) {
			results = append(results, record)
		}
	}

	// Offset適用
	if query.Offset > 0 && query.Offset < len(results) {
		results = results[query.Offset:]
	} else if query.Offset >= len(results) {
		return []*Record{}
	}

	// Limit適用
	if query.Limit > 0 && query.Limit < len(results) {
		results = results[:query.Limit]
	}

	return results
}

// ValidateQuery validates query structure
func ValidateQuery(query Query) error {
	for i, cond := range query.Conditions {
		// 演算子の検証
		validOps := []string{"==", "!=", ">", ">=", "<", "<=", "in", "between"}
		valid := false
		for _, op := range validOps {
			if cond.Operator == op {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid operator '%s' in condition %d", cond.Operator, i)
		}

		// in演算子の値検証
		if cond.Operator == "in" {
			if _, ok := cond.Value.([]interface{}); !ok {
				return fmt.Errorf("operator 'in' requires []interface{} value in condition %d", i)
			}
		}

		// between演算子の値検証
		if cond.Operator == "between" {
			valid := false
			switch v := cond.Value.(type) {
			case [2]interface{}:
				valid = true
			case []interface{}:
				if len(v) == 2 {
					valid = true
				}
			}
			if !valid {
				return fmt.Errorf("operator 'between' requires [2]interface{} or []interface{} with 2 elements in condition %d", i)
			}
		}

		// カラム名の検証
		if cond.Column == "" {
			return fmt.Errorf("empty column name in condition %d", i)
		}
	}

	// Limit/Offsetの検証
	if query.Limit < 0 {
		return fmt.Errorf("limit must be non-negative")
	}
	if query.Offset < 0 {
		return fmt.Errorf("offset must be non-negative")
	}

	return nil
}
