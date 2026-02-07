package email

import "strings"

// CanonicalEmail 1. Trim spaces 2. Convert to lower case
func CanonicalEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
