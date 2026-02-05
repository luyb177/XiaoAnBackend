package email

import (
	"regexp"
)

var emailRegexp = regexp.MustCompile(
	`^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$`,
)

func IsValidEmail(email string) bool {
	return emailRegexp.MatchString(email)
}
