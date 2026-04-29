package services

import (
	"testing"
)

func TestRandomGenerator_LengthAndAlphabet(t *testing.T) {
	g := NewRandomGenerator()
	allowed := make(map[byte]struct{}, len(alphabet))
	for i := 0; i < len(alphabet); i++ {
		allowed[alphabet[i]] = struct{}{}
	}

	for i := 0; i < 1000; i++ {
		s := g.Generate()
		if len(s) != ShortURLLength {
			t.Fatalf("len=%d, want %d", len(s), ShortURLLength)
		}
		for j := 0; j < len(s); j++ {
			if _, ok := allowed[s[j]]; !ok {
				t.Fatalf("char %q at %d not in alphabet", s[j], j)
			}
		}
	}
}

func TestRandomGenerator_Uniqueness(t *testing.T) {
	g := NewRandomGenerator()
	seen := make(map[string]struct{}, 1000)
	for i := 0; i < 1000; i++ {
		s := g.Generate()
		if _, dup := seen[s]; dup {
			t.Fatalf("duplicate generated: %q (collision in 1000 samples is extremely unlikely)", s)
		}
		seen[s] = struct{}{}
	}
}
