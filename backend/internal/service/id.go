package service

import (
	"crypto/rand"
	"encoding/hex"
)

func NewProjectID() string {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		panic(err)
	}
	return hex.EncodeToString(raw[:])
}
