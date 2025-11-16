package controllers

import (
	"crypto/rand"
	"encoding/base32"
)

func (Application) generateRandomString() string {
	b := make([]byte, 16)
	rand.Read(b)

	s := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)

	return s
}
