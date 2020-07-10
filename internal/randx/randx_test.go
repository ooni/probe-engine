package randx_test

import (
	"testing"

	"github.com/ooni/probe-engine/internal/randx"
)

func TestLetters(t *testing.T) {
	str := randx.Letters(1024)
	for _, chr := range str {
		if (chr >= 'A' && chr <= 'Z') || (chr >= 'a' && chr <= 'z') {
			continue
		}
		t.Fatal("invalid input char")
	}
}

func TestLettersUppercase(t *testing.T) {
	str := randx.LettersUppercase(1024)
	for _, chr := range str {
		if chr >= 'A' && chr <= 'Z' {
			continue
		}
		t.Fatal("invalid input char")
	}
}
