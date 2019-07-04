// Package onionkey manages onion service key generation, serialization, and
// service ID calculation. Currently only supports version 3 onions.
package onionkey

import (
	"github.com/wybiral/torgo"
)

// Key is generic interface type for Tor onion keys.
type Key interface {
	WriteFile(path string) error
	Onion() (*torgo.Onion, error)
	ServiceID() string
}

// GenerateKey generates a Tor onion key.
func GenerateKey() (Key, error) {
	return generateV3()
}

// ReadFile reads a Tor onion key from file path.
func ReadFile(path string) (Key, error) {
	return readV3(path)
}
