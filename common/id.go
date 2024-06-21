package common

import "math/rand/v2"

// ID of a session.
type ID uint32

func NewID() ID {
	for {
		id := ID(rand.Uint32())
		if id != 0 {
			return id
		}
	}
}
