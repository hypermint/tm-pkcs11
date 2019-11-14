package helpers

import "math/rand"

func RandomBytes32() []byte {
	result := make([]byte, 32)
	rand.Read(result)
	return result
}
