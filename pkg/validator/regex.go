package validator

import "regexp"

var (
	EmailRX    = regexp.MustCompile(`^([a-zA-Z0-9._%-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,6})*$`)
	LowerRX    = regexp.MustCompile(`[a-z]`)
	AlphanumRX = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	UpperRX    = regexp.MustCompile(`[A-Z]`)
	NumberRX   = regexp.MustCompile(`[0-9]`)
	SpecialRX  = regexp.MustCompile(`[!@#$%^&*]`)
)
