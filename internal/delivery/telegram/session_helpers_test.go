package telegram

import "testing"

func TestSessionString(t *testing.T) {
	tests := []struct {
		name string
		s    map[string]interface{}
		key  string
		want string
	}{
		{"present string", map[string]interface{}{"k": "v"}, "k", "v"},
		{"empty string value", map[string]interface{}{"k": ""}, "k", ""},
		{"missing key", map[string]interface{}{}, "k", ""},
		{"nil map", nil, "k", ""},
		{"wrong type int", map[string]interface{}{"k": 42}, "k", ""},
		{"wrong type bool", map[string]interface{}{"k": true}, "k", ""},
		{"wrong type slice", map[string]interface{}{"k": []string{"a"}}, "k", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sessionString(tt.s, tt.key)
			if got != tt.want {
				t.Errorf("sessionString: got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSessionBool(t *testing.T) {
	tests := []struct {
		name string
		s    map[string]interface{}
		key  string
		want bool
	}{
		{"true", map[string]interface{}{"k": true}, "k", true},
		{"false", map[string]interface{}{"k": false}, "k", false},
		{"missing key", map[string]interface{}{}, "k", false},
		{"nil map", nil, "k", false},
		{"wrong type string", map[string]interface{}{"k": "true"}, "k", false},
		{"wrong type int", map[string]interface{}{"k": 1}, "k", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sessionBool(tt.s, tt.key)
			if got != tt.want {
				t.Errorf("sessionBool: got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSessionHasKey(t *testing.T) {
	tests := []struct {
		name string
		s    map[string]interface{}
		key  string
		want bool
	}{
		{"present", map[string]interface{}{"k": "anything"}, "k", true},
		{"present with nil value", map[string]interface{}{"k": nil}, "k", true},
		{"missing", map[string]interface{}{}, "k", false},
		{"nil map", nil, "k", false},
		{"different key present", map[string]interface{}{"other": 1}, "k", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sessionHasKey(tt.s, tt.key)
			if got != tt.want {
				t.Errorf("sessionHasKey: got %v, want %v", got, tt.want)
			}
		})
	}
}
