package torgo

import (
	"crypto/sha512"
	"encoding/base32"
	"encoding/base64"
	"errors"
	"strings"

	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/sha3"
)

// OnionFromEd25519 returns an Onion instance from an ED25519 private key which
// can be used to start a hidden service with controller.AddOnion.
func OnionFromEd25519(pri ed25519.PrivateKey) (*Onion, error) {
	pub, ok := pri.Public().(ed25519.PublicKey)
	if !ok {
		return nil, errors.New("torgo: unable to extract ed25519.PublicKey")
	}
	serviceID, err := ServiceIDFromEd25519(pub)
	if err != nil {
		return nil, err
	}
	h := sha512.Sum512(pri[:32])
	// Set bits so that h[:32] is private scalar "a"
	h[0] &= 248
	h[31] &= 127
	h[31] |= 64
	// Since h[32:] is RH, h is now (a || RH)
	key := base64.StdEncoding.EncodeToString(h[:])
	return &Onion{
		Ports:          make(map[int]string),
		ServiceID:      serviceID,
		PrivateKey:     key,
		PrivateKeyType: "ED25519-V3",
	}, nil
}

// ServiceIDFromEd25519 calculates a Tor service ID from an ed25519.PublicKey.
func ServiceIDFromEd25519(pub ed25519.PublicKey) (string, error) {
	checkdigits := ed25519Checkdigits(pub)
	combined := pub[:]
	combined = append(combined, checkdigits...)
	combined = append(combined, 0x03)
	serviceID := base32.StdEncoding.EncodeToString(combined)
	return strings.ToLower(serviceID), nil
}

// ed25519Checkdigits calculates tje check digits used to create service ID.
func ed25519Checkdigits(pub ed25519.PublicKey) []byte {
	checkstr := []byte(".onion checksum")
	checkstr = append(checkstr, pub...)
	checkstr = append(checkstr, 0x03)
	checksum := sha3.Sum256(checkstr)
	return checksum[:2]
}
