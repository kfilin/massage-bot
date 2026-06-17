package config

// ResolveAdminIDs merges the primary admin, allowed-telegram-IDs list, and
// therapist IDs into a single deduplicated slice. The result preserves no
// particular order (map iteration is random in Go), which is fine because
// callers either iterate the slice for notifications or check membership
// via IsAdmin.
func ResolveAdminIDs(adminPrimary string, allowedIDs []string, therapistIDs []string) []string {
	seen := make(map[string]struct{})
	if adminPrimary != "" {
		seen[adminPrimary] = struct{}{}
	}
	for _, id := range allowedIDs {
		if id != "" {
			seen[id] = struct{}{}
		}
	}
	for _, id := range therapistIDs {
		if id != "" {
			seen[id] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for id := range seen {
		out = append(out, id)
	}
	return out
}
