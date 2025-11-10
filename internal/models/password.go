package models

import "golang.org/x/crypto/bcrypt"

func NewPassword(password string) (*Password, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return &Password{
		password: &password,
		hash:     hash,
	}, nil
}

type Password struct {
	password *string
	hash     []byte
}

func (p *Password) ComparePassword(password string) bool {
	return bcrypt.CompareHashAndPassword(p.hash, []byte(password)) == nil
}
