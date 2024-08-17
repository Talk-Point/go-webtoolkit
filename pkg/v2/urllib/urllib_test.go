package urllib

import (
	"testing"
)

func TestAddParam(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    interface{}
		expected string
	}{
		{"AddBoolTrue", "boolKey", true, "true"},
		{"AddBoolFalse", "boolKey", false, "false"},
		{"AddString", "stringKey", "stringValue", "stringValue"},
		{"AddInt", "intKey", 123, "123"},
		{"AddFloat", "floatKey", 123.456, "123.456000"},
		{"AddUnsupportedType", "unsupportedKey", []int{1, 2, 3}, "[1 2 3]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseUrl := Url("http://example.com")
			baseUrl.AddParam(tt.key, tt.value)
			if baseUrl.Params[tt.key] != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, baseUrl.Params[tt.key])
			}
		})
	}
}

func TestUrlParams(t *testing.T) {
	tests := []struct {
		base     string
		params   map[string]interface{}
		expected string
	}{
		{
			base: "/api/:mandant",
			params: map[string]interface{}{
				"mandant": 1,
			},
			expected: "/api/1",
		},
		{
			base: "/api/:mandant",
			params: map[string]interface{}{
				"mandant":    1,
				"otherParam": "value",
			},
			expected: "/api/1",
		},
		{
			base:     "/api/:mandant",
			params:   nil,
			expected: "/api/:mandant",
		},
		{
			base: "/api/:value/:value",
			params: map[string]interface{}{
				"value": 1,
			},
			expected: "/api/1/1",
		},
		{
			base: "/api/:mandant",
			params: map[string]interface{}{
				"mandant":    true,
				"boolParam":  true,
				"intParam":   42,
				"floatParam": 3.14,
			},
			expected: "/api/true",
		},
	}

	for _, test := range tests {
		url := Url(test.base, test.params)
		if url.String() != test.expected {
			t.Errorf("for base %s with params %v, expected %s, got %s", test.base, test.params, test.expected, url.String())
		}
	}
}
