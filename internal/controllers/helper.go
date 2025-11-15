package controllers

import "crypto/rand"

func (Controller) generateRandomString() string {
	b := make([]byte, 16)
	rand.Read(b)

	return string(b)
}
