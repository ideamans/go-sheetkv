package sheetkv_test

import (
	"testing"

	"github.com/ideamans/go-sheetkv"
)

func TestRecord_MatchesQuery(t *testing.T) {
	tests := []struct {
		name   string
		record sheetkv.Record
		query  sheetkv.Query
		want   bool
	}{
		{
			name: "single == condition match",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"status": "active"},
			},
			query: sheetkv.Query{
				Conditions: []sheetkv.Condition{
					{Column: "status", Operator: "==", Value: "active"},
				},
			},
			want: true,
		},
		{
			name: "single == condition no match",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"status": "inactive"},
			},
			query: sheetkv.Query{
				Conditions: []sheetkv.Condition{
					{Column: "status", Operator: "==", Value: "active"},
				},
			},
			want: false,
		},
		{
			name: "!= condition",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"status": "inactive"},
			},
			query: sheetkv.Query{
				Conditions: []sheetkv.Condition{
					{Column: "status", Operator: "!=", Value: "active"},
				},
			},
			want: true,
		},
		{
			name: "> condition with numbers",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"age": 25},
			},
			query: sheetkv.Query{
				Conditions: []sheetkv.Condition{
					{Column: "age", Operator: ">", Value: 20},
				},
			},
			want: true,
		},
		{
			name: ">= condition with equal values",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"age": 20},
			},
			query: sheetkv.Query{
				Conditions: []sheetkv.Condition{
					{Column: "age", Operator: ">=", Value: 20},
				},
			},
			want: true,
		},
		{
			name: "< condition",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"age": 15},
			},
			query: sheetkv.Query{
				Conditions: []sheetkv.Condition{
					{Column: "age", Operator: "<", Value: 20},
				},
			},
			want: true,
		},
		{
			name: "<= condition",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"age": 20},
			},
			query: sheetkv.Query{
				Conditions: []sheetkv.Condition{
					{Column: "age", Operator: "<=", Value: 20},
				},
			},
			want: true,
		},
		{
			name: "in condition match",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"role": "admin"},
			},
			query: sheetkv.Query{
				Conditions: []sheetkv.Condition{
					{Column: "role", Operator: "in", Value: []interface{}{"admin", "moderator", "user"}},
				},
			},
			want: true,
		},
		{
			name: "in condition no match",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"role": "guest"},
			},
			query: sheetkv.Query{
				Conditions: []sheetkv.Condition{
					{Column: "role", Operator: "in", Value: []interface{}{"admin", "moderator", "user"}},
				},
			},
			want: false,
		},
		{
			name: "between condition match",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"age": 25},
			},
			query: sheetkv.Query{
				Conditions: []sheetkv.Condition{
					{Column: "age", Operator: "between", Value: [2]interface{}{20, 30}},
				},
			},
			want: true,
		},
		{
			name: "between condition with slice",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"age": 25},
			},
			query: sheetkv.Query{
				Conditions: []sheetkv.Condition{
					{Column: "age", Operator: "between", Value: []interface{}{20, 30}},
				},
			},
			want: true,
		},
		{
			name: "between condition no match",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"age": 35},
			},
			query: sheetkv.Query{
				Conditions: []sheetkv.Condition{
					{Column: "age", Operator: "between", Value: [2]interface{}{20, 30}},
				},
			},
			want: false,
		},
		{
			name: "multiple conditions AND",
			record: sheetkv.Record{
				Key: 2,
				Values: map[string]interface{}{
					"age":    25,
					"status": "active",
					"role":   "admin",
				},
			},
			query: sheetkv.Query{
				Conditions: []sheetkv.Condition{
					{Column: "age", Operator: ">=", Value: 20},
					{Column: "status", Operator: "==", Value: "active"},
					{Column: "role", Operator: "in", Value: []interface{}{"admin", "moderator"}},
				},
			},
			want: true,
		},
		{
			name: "multiple conditions one fails",
			record: sheetkv.Record{
				Key: 2,
				Values: map[string]interface{}{
					"age":    25,
					"status": "inactive",
					"role":   "admin",
				},
			},
			query: sheetkv.Query{
				Conditions: []sheetkv.Condition{
					{Column: "age", Operator: ">=", Value: 20},
					{Column: "status", Operator: "==", Value: "active"},
					{Column: "role", Operator: "in", Value: []interface{}{"admin", "moderator"}},
				},
			},
			want: false,
		},
		{
			name: "missing column treated as null",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"name": "John"},
			},
			query: sheetkv.Query{
				Conditions: []sheetkv.Condition{
					{Column: "age", Operator: "==", Value: nil},
				},
			},
			want: true,
		},
		{
			name: "compare different numeric types",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"count": int64(100)},
			},
			query: sheetkv.Query{
				Conditions: []sheetkv.Condition{
					{Column: "count", Operator: "==", Value: 100.0},
				},
			},
			want: true,
		},
		{
			name: "invalid operator",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"status": "active"},
			},
			query: sheetkv.Query{
				Conditions: []sheetkv.Condition{
					{Column: "status", Operator: "invalid", Value: "active"},
				},
			},
			want: false,
		},
		{
			name: "empty conditions matches all",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"status": "active"},
			},
			query: sheetkv.Query{
				Conditions: []sheetkv.Condition{},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.record.MatchesQuery(tt.query)
			if got != tt.want {
				t.Errorf("MatchesQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApplyQuery(t *testing.T) {
	records := []*sheetkv.Record{
		{
			Key:    2,
			Values: map[string]interface{}{"age": 25, "status": "active"},
		},
		{
			Key:    3,
			Values: map[string]interface{}{"age": 30, "status": "inactive"},
		},
		{
			Key:    4,
			Values: map[string]interface{}{"age": 35, "status": "active"},
		},
		{
			Key:    5,
			Values: map[string]interface{}{"age": 20, "status": "active"},
		},
		{
			Key:    6,
			Values: map[string]interface{}{"age": 40, "status": "inactive"},
		},
	}

	tests := []struct {
		name    string
		records []*sheetkv.Record
		query   sheetkv.Query
		want    []int // expected keys
	}{
		{
			name:    "filter by status",
			records: records,
			query: sheetkv.Query{
				Conditions: []sheetkv.Condition{
					{Column: "status", Operator: "==", Value: "active"},
				},
			},
			want: []int{2, 4, 5},
		},
		{
			name:    "filter by age range",
			records: records,
			query: sheetkv.Query{
				Conditions: []sheetkv.Condition{
					{Column: "age", Operator: "between", Value: [2]interface{}{25, 35}},
				},
			},
			want: []int{2, 3, 4},
		},
		{
			name:    "with limit",
			records: records,
			query: sheetkv.Query{
				Conditions: []sheetkv.Condition{
					{Column: "status", Operator: "==", Value: "active"},
				},
				Limit: 2,
			},
			want: []int{2, 4},
		},
		{
			name:    "with offset",
			records: records,
			query: sheetkv.Query{
				Conditions: []sheetkv.Condition{
					{Column: "status", Operator: "==", Value: "active"},
				},
				Offset: 1,
			},
			want: []int{4, 5},
		},
		{
			name:    "with limit and offset",
			records: records,
			query: sheetkv.Query{
				Conditions: []sheetkv.Condition{
					{Column: "status", Operator: "==", Value: "active"},
				},
				Limit:  1,
				Offset: 1,
			},
			want: []int{4},
		},
		{
			name:    "offset beyond results",
			records: records,
			query: sheetkv.Query{
				Conditions: []sheetkv.Condition{
					{Column: "status", Operator: "==", Value: "active"},
				},
				Offset: 10,
			},
			want: []int{},
		},
		{
			name:    "no conditions",
			records: records,
			query: sheetkv.Query{
				Limit: 3,
			},
			want: []int{2, 3, 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sheetkv.ApplyQuery(tt.records, tt.query)
			if len(got) != len(tt.want) {
				t.Errorf("ApplyQuery() returned %d records, want %d", len(got), len(tt.want))
				return
			}
			for i, record := range got {
				if record.Key != tt.want[i] {
					t.Errorf("ApplyQuery()[%d].Key = %v, want %v", i, record.Key, tt.want[i])
				}
			}
		})
	}
}

func TestValidateQuery(t *testing.T) {
	tests := []struct {
		name    string
		query   sheetkv.Query
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid query",
			query: sheetkv.Query{
				Conditions: []sheetkv.Condition{
					{Column: "status", Operator: "==", Value: "active"},
					{Column: "age", Operator: ">=", Value: 20},
				},
				Limit:  10,
				Offset: 0,
			},
			wantErr: false,
		},
		{
			name: "invalid operator",
			query: sheetkv.Query{
				Conditions: []sheetkv.Condition{
					{Column: "status", Operator: "invalid", Value: "active"},
				},
			},
			wantErr: true,
			errMsg:  "invalid operator",
		},
		{
			name: "in operator with non-slice value",
			query: sheetkv.Query{
				Conditions: []sheetkv.Condition{
					{Column: "role", Operator: "in", Value: "admin"},
				},
			},
			wantErr: true,
			errMsg:  "operator 'in' requires []interface{}",
		},
		{
			name: "between operator with invalid value",
			query: sheetkv.Query{
				Conditions: []sheetkv.Condition{
					{Column: "age", Operator: "between", Value: 25},
				},
			},
			wantErr: true,
			errMsg:  "operator 'between' requires",
		},
		{
			name: "between operator with wrong slice length",
			query: sheetkv.Query{
				Conditions: []sheetkv.Condition{
					{Column: "age", Operator: "between", Value: []interface{}{20}},
				},
			},
			wantErr: true,
			errMsg:  "operator 'between' requires",
		},
		{
			name: "empty column name",
			query: sheetkv.Query{
				Conditions: []sheetkv.Condition{
					{Column: "", Operator: "==", Value: "active"},
				},
			},
			wantErr: true,
			errMsg:  "empty column name",
		},
		{
			name: "negative limit",
			query: sheetkv.Query{
				Conditions: []sheetkv.Condition{},
				Limit:      -1,
			},
			wantErr: true,
			errMsg:  "limit must be non-negative",
		},
		{
			name: "negative offset",
			query: sheetkv.Query{
				Conditions: []sheetkv.Condition{},
				Offset:     -1,
			},
			wantErr: true,
			errMsg:  "offset must be non-negative",
		},
		{
			name: "valid in operator",
			query: sheetkv.Query{
				Conditions: []sheetkv.Condition{
					{Column: "role", Operator: "in", Value: []interface{}{"admin", "user"}},
				},
			},
			wantErr: false,
		},
		{
			name: "valid between operator with array",
			query: sheetkv.Query{
				Conditions: []sheetkv.Condition{
					{Column: "age", Operator: "between", Value: [2]interface{}{20, 30}},
				},
			},
			wantErr: false,
		},
		{
			name: "valid between operator with slice",
			query: sheetkv.Query{
				Conditions: []sheetkv.Condition{
					{Column: "age", Operator: "between", Value: []interface{}{20, 30}},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sheetkv.ValidateQuery(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateQuery() error = %v, want error containing %v", err, tt.errMsg)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || len(s) > len(substr) && contains(s[1:], substr)
}
