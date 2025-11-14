package validator

import "regexp"

var (
	EmailRX    = regexp.MustCompile(`^([a-zA-Z0-9._%-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,6})*$`)
	PasswordRX = regexp.MustCompile(`^[A-Za-z0-9!@#\$%\^&\*\(\)\-_=+\[\]{};:'",.<>\/?\\|` + "`" + `~]{8,50}$`)
)
