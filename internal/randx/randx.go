// Package randx contains math/rand extensions
package randx

import (
	"math/rand"
	"time"
)

const (
	uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lowercase = "abcdefghijklmnopqrstuvwxyz"
	letters   = uppercase + lowercase
)

func lettersWithString(n int, letterBytes string) string {
	// See https://stackoverflow.com/questions/22892120
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rnd.Intn(len(letterBytes))]
	}
	return string(b)
}

// Letters return a string composed of random letters
func Letters(n int) string {
	return lettersWithString(n, letters)
}

// LettersUppercase return a string composed of random uppercase letters
func LettersUppercase(n int) string {
	return lettersWithString(n, uppercase)
}
