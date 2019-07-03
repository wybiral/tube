package onionkey

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"os"
	"strings"

	"github.com/wybiral/torgo"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/sha3"
)

type v3Key ed25519.PrivateKey

func generateV3() (v3Key, error) {
	_, key, err := ed25519.GenerateKey(rand.Reader)
	return v3Key(key), err
}

func readV3(path string) (v3Key, error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	pk := strings.TrimSpace(string(raw))
	parts := strings.SplitN(pk, ":", 2)
	if parts[0] != "v3" {
		return nil, errors.New("Invalid key type")
	}
	seed, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	key := ed25519.NewKeyFromSeed(seed)
	return v3Key(key), nil
}

func (k v3Key) Onion() (*torgo.Onion, error) {
	return torgo.OnionFromEd25519(ed25519.PrivateKey(k))
}

func (k v3Key) WriteFile(path string) error {
	seed := ed25519.PrivateKey(k).Seed()
	b64 := base64.StdEncoding.EncodeToString(seed)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString("v3:" + b64)
	if err != nil {
		return err
	}
	return nil
}

func (k v3Key) ServiceID() string {
	// Get ed25519 public key
	pub := ed25519.PrivateKey(k).Public().(ed25519.PublicKey)
	// Calculate check digits
	checkstr := []byte(".onion checksum")
	checkstr = append(checkstr, pub...)
	checkstr = append(checkstr, 0x03)
	checksum := sha3.Sum256(checkstr)
	checkdigits := checksum[:2]
	// Calculate service ID
	combined := pub[:]
	combined = append(combined, checkdigits...)
	combined = append(combined, 0x03)
	serviceID := base32.StdEncoding.EncodeToString(combined)
	return strings.ToLower(serviceID)
}
