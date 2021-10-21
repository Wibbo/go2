package utils

import (
	"math/rand"
)

// PlusOrMinus Randomly returns either -1 or +1.
func PlusOrMinus() int {
	return 2*rand.Intn(2)-1
}