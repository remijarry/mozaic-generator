package noise

import (
	"math/rand"
	"time"
)

// Perlin noise functions will be implemented here.
func Generate(x, y int) float32 {
	rand.Seed(time.Now().UnixNano())

	return rand.Float32()
}
