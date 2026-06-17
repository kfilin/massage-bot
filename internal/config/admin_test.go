package config

import (
	"reflect"
	"testing"
)

func TestResolveAdminIDs(t *testing.T) {
	tests := []struct {
		name     string
		primary  string
		allowed  []string
		therapist []string
		want     []string
	}{
		{
			name:    "all empty",
			want:    []string{},
		},
		{
			name:    "only primary",
			primary: "111",
			want:    []string{"111"},
		},
		{
			name:    "dedup across lists",
			primary: "111",
			allowed: []string{"111", "222", ""},
			therapist: []string{"222", "333"},
			// map iteration is random, so we can't assert exact order
			want: nil, // checked below
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveAdminIDs(tt.primary, tt.allowed, tt.therapist)
			if tt.want == nil {
				// Dedup test: just check the set
				if len(got) != 3 {
					t.Errorf("expected 3 unique IDs, got %d: %v", len(got), got)
				}
				set := make(map[string]struct{})
				for _, id := range got {
					set[id] = struct{}{}
				}
				for _, want := range []string{"111", "222", "333"} {
					if _, ok := set[want]; !ok {
						t.Errorf("missing %s in %v", want, got)
					}
				}
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ResolveAdminIDs = %v, want %v", got, tt.want)
			}
		})
	}
}
