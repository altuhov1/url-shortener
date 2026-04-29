package services

import (
	"math/rand/v2"
	"strings"
)

const (
	ShortURLLength = 10
	alphabet       = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
)

type RandomGenerator struct{}

func NewRandomGenerator() *RandomGenerator {
	return &RandomGenerator{}
}

func (g *RandomGenerator) Generate() string {
	var sb strings.Builder
	sb.Grow(ShortURLLength)
	for i := 0; i < ShortURLLength; i++ {
		sb.WriteByte(alphabet[rand.IntN(len(alphabet))])
	}
	return sb.String()
}
