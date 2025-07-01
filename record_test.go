package sheetkv_test

import (
	"testing"
	"time"

	"github.com/ideamans/go-sheetkv"
)

func TestRecord_GetAsString(t *testing.T) {
	tests := []struct {
		name         string
		record       sheetkv.Record
		col          string
		defaultValue string
		want         string
	}{
		{
			name: "string value",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"name": "John Doe"},
			},
			col:          "name",
			defaultValue: "default",
			want:         "John Doe",
		},
		{
			name: "int value",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"age": 30},
			},
			col:          "age",
			defaultValue: "default",
			want:         "30",
		},
		{
			name: "float64 value",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"score": 99.5},
			},
			col:          "score",
			defaultValue: "default",
			want:         "99.5",
		},
		{
			name: "bool true",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"active": true},
			},
			col:          "active",
			defaultValue: "default",
			want:         "true",
		},
		{
			name: "bool false",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"active": false},
			},
			col:          "active",
			defaultValue: "default",
			want:         "false",
		},
		{
			name: "[]string value",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"tags": []string{"tag1", "tag2", "tag3"}},
			},
			col:          "tags",
			defaultValue: "default",
			want:         "tag1,tag2,tag3",
		},
		{
			name: "missing value",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{},
			},
			col:          "missing",
			defaultValue: "default",
			want:         "default",
		},
		{
			name: "nil value",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"nullval": nil},
			},
			col:          "nullval",
			defaultValue: "default",
			want:         "<nil>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.record.GetAsString(tt.col, tt.defaultValue)
			if got != tt.want {
				t.Errorf("GetAsString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRecord_GetAsInt64(t *testing.T) {
	tests := []struct {
		name         string
		record       sheetkv.Record
		col          string
		defaultValue int64
		want         int64
	}{
		{
			name: "int64 value",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"count": int64(100)},
			},
			col:          "count",
			defaultValue: -1,
			want:         100,
		},
		{
			name: "int value",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"count": 100},
			},
			col:          "count",
			defaultValue: -1,
			want:         100,
		},
		{
			name: "float64 value",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"count": 100.5},
			},
			col:          "count",
			defaultValue: -1,
			want:         100,
		},
		{
			name: "string numeric value",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"count": "100"},
			},
			col:          "count",
			defaultValue: -1,
			want:         100,
		},
		{
			name: "string non-numeric value",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"count": "abc"},
			},
			col:          "count",
			defaultValue: -1,
			want:         -1,
		},
		{
			name: "missing value",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{},
			},
			col:          "missing",
			defaultValue: -1,
			want:         -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.record.GetAsInt64(tt.col, tt.defaultValue)
			if got != tt.want {
				t.Errorf("GetAsInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRecord_GetAsFloat64(t *testing.T) {
	tests := []struct {
		name         string
		record       sheetkv.Record
		col          string
		defaultValue float64
		want         float64
	}{
		{
			name: "float64 value",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"score": 99.5},
			},
			col:          "score",
			defaultValue: -1.0,
			want:         99.5,
		},
		{
			name: "int value",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"score": 100},
			},
			col:          "score",
			defaultValue: -1.0,
			want:         100.0,
		},
		{
			name: "int64 value",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"score": int64(100)},
			},
			col:          "score",
			defaultValue: -1.0,
			want:         100.0,
		},
		{
			name: "string numeric value",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"score": "99.5"},
			},
			col:          "score",
			defaultValue: -1.0,
			want:         99.5,
		},
		{
			name: "string non-numeric value",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"score": "abc"},
			},
			col:          "score",
			defaultValue: -1.0,
			want:         -1.0,
		},
		{
			name: "missing value",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{},
			},
			col:          "missing",
			defaultValue: -1.0,
			want:         -1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.record.GetAsFloat64(tt.col, tt.defaultValue)
			if got != tt.want {
				t.Errorf("GetAsFloat64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRecord_GetAsStrings(t *testing.T) {
	tests := []struct {
		name         string
		record       sheetkv.Record
		col          string
		defaultValue []string
		want         []string
	}{
		{
			name: "[]string value",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"tags": []string{"tag1", "tag2", "tag3"}},
			},
			col:          "tags",
			defaultValue: []string{"default"},
			want:         []string{"tag1", "tag2", "tag3"},
		},
		{
			name: "comma-separated string",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"tags": "tag1,tag2,tag3"},
			},
			col:          "tags",
			defaultValue: []string{"default"},
			want:         []string{"tag1", "tag2", "tag3"},
		},
		{
			name: "empty string",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"tags": ""},
			},
			col:          "tags",
			defaultValue: []string{"default"},
			want:         []string{},
		},
		{
			name: "[]interface{} value",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"tags": []interface{}{"tag1", 2, true}},
			},
			col:          "tags",
			defaultValue: []string{"default"},
			want:         []string{"tag1", "2", "true"},
		},
		{
			name: "missing value",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{},
			},
			col:          "missing",
			defaultValue: []string{"default"},
			want:         []string{"default"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.record.GetAsStrings(tt.col, tt.defaultValue)
			if len(got) != len(tt.want) {
				t.Errorf("GetAsStrings() len = %v, want %v", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("GetAsStrings()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestRecord_GetAsBool(t *testing.T) {
	tests := []struct {
		name         string
		record       sheetkv.Record
		col          string
		defaultValue bool
		want         bool
	}{
		{
			name: "bool true",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"active": true},
			},
			col:          "active",
			defaultValue: false,
			want:         true,
		},
		{
			name: "bool false",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"active": false},
			},
			col:          "active",
			defaultValue: true,
			want:         false,
		},
		{
			name: "string true",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"active": "true"},
			},
			col:          "active",
			defaultValue: false,
			want:         true,
		},
		{
			name: "string 1",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"active": "1"},
			},
			col:          "active",
			defaultValue: false,
			want:         true,
		},
		{
			name: "string false",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"active": "false"},
			},
			col:          "active",
			defaultValue: true,
			want:         false,
		},
		{
			name: "int 1",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"active": 1},
			},
			col:          "active",
			defaultValue: false,
			want:         true,
		},
		{
			name: "int 0",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"active": 0},
			},
			col:          "active",
			defaultValue: true,
			want:         false,
		},
		{
			name: "float64 non-zero",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"active": 1.5},
			},
			col:          "active",
			defaultValue: false,
			want:         true,
		},
		{
			name: "float64 zero",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"active": 0.0},
			},
			col:          "active",
			defaultValue: true,
			want:         false,
		},
		{
			name: "missing value",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{},
			},
			col:          "missing",
			defaultValue: true,
			want:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.record.GetAsBool(tt.col, tt.defaultValue)
			if got != tt.want {
				t.Errorf("GetAsBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRecord_GetAsTime(t *testing.T) {
	rfc3339Time, _ := time.Parse(time.RFC3339, "2023-12-25T12:00:00Z")
	customTime, _ := time.Parse("2006-01-02 15:04:05", "2023-12-25 12:00:00")
	dateOnlyTime, _ := time.Parse("2006-01-02", "2023-12-25")
	defaultTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		record       sheetkv.Record
		col          string
		defaultValue time.Time
		want         time.Time
	}{
		{
			name: "time.Time value",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"created": rfc3339Time},
			},
			col:          "created",
			defaultValue: defaultTime,
			want:         rfc3339Time,
		},
		{
			name: "RFC3339 string",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"created": "2023-12-25T12:00:00Z"},
			},
			col:          "created",
			defaultValue: defaultTime,
			want:         rfc3339Time,
		},
		{
			name: "custom format string",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"created": "2023-12-25 12:00:00"},
			},
			col:          "created",
			defaultValue: defaultTime,
			want:         customTime,
		},
		{
			name: "date only string",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"created": "2023-12-25"},
			},
			col:          "created",
			defaultValue: defaultTime,
			want:         dateOnlyTime,
		},
		{
			name: "invalid string",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{"created": "invalid"},
			},
			col:          "created",
			defaultValue: defaultTime,
			want:         defaultTime,
		},
		{
			name: "missing value",
			record: sheetkv.Record{
				Key:    2,
				Values: map[string]interface{}{},
			},
			col:          "missing",
			defaultValue: defaultTime,
			want:         defaultTime,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.record.GetAsTime(tt.col, tt.defaultValue)
			if !got.Equal(tt.want) {
				t.Errorf("GetAsTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRecord_Setters(t *testing.T) {
	t.Run("SetString", func(t *testing.T) {
		r := &sheetkv.Record{Key: 2}
		r.SetString("name", "John Doe")
		if r.Values["name"] != "John Doe" {
			t.Errorf("SetString() failed, got %v", r.Values["name"])
		}
	})

	t.Run("SetInt64", func(t *testing.T) {
		r := &sheetkv.Record{Key: 2}
		r.SetInt64("age", 30)
		if r.Values["age"] != int64(30) {
			t.Errorf("SetInt64() failed, got %v", r.Values["age"])
		}
	})

	t.Run("SetFloat64", func(t *testing.T) {
		r := &sheetkv.Record{Key: 2}
		r.SetFloat64("score", 99.5)
		if r.Values["score"] != 99.5 {
			t.Errorf("SetFloat64() failed, got %v", r.Values["score"])
		}
	})

	t.Run("SetStrings", func(t *testing.T) {
		r := &sheetkv.Record{Key: 2}
		r.SetStrings("tags", []string{"tag1", "tag2", "tag3"})
		if r.Values["tags"] != "tag1,tag2,tag3" {
			t.Errorf("SetStrings() failed, got %v", r.Values["tags"])
		}
	})

	t.Run("SetBool", func(t *testing.T) {
		r := &sheetkv.Record{Key: 2}
		r.SetBool("active", true)
		if r.Values["active"] != true {
			t.Errorf("SetBool() failed, got %v", r.Values["active"])
		}
	})

	t.Run("SetTime", func(t *testing.T) {
		r := &sheetkv.Record{Key: 2}
		testTime := time.Date(2023, 12, 25, 12, 0, 0, 0, time.UTC)
		r.SetTime("created", testTime)
		expected := "2023-12-25T12:00:00Z"
		if r.Values["created"] != expected {
			t.Errorf("SetTime() failed, got %v, want %v", r.Values["created"], expected)
		}
	})

	t.Run("SetString on nil Values", func(t *testing.T) {
		r := &sheetkv.Record{Key: 2, Values: nil}
		r.SetString("name", "John Doe")
		if r.Values == nil {
			t.Error("SetString() should initialize Values map")
		}
		if r.Values["name"] != "John Doe" {
			t.Errorf("SetString() failed, got %v", r.Values["name"])
		}
	})
}
