package torgo

import (
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base32"
	"encoding/base64"
	"errors"
	"strings"
)

// OnionFromRSA returns an Onion instance from a 1024 bit RSA private key which
// can be used to start a hidden service with controller.AddOnion.
func OnionFromRSA(pri *rsa.PrivateKey) (*Onion, error) {
	pub, ok := pri.Public().(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("torgo: unable to extract *rsa.PublicKey")
	}
	serviceID, err := ServiceIDFromRSA(pub)
	if err != nil {
		return nil, err
	}
	der := x509.MarshalPKCS1PrivateKey(pri)
	key := base64.StdEncoding.EncodeToString(der)
	return &Onion{
		Ports:          make(map[int]string),
		ServiceID:      serviceID,
		PrivateKey:     key,
		PrivateKeyType: "RSA1024",
	}, nil
}

// ServiceIDFromRSA calculates a Tor service ID from an *rsa.PublicKey.
func ServiceIDFromRSA(pub *rsa.PublicKey) (string, error) {
	der, err := asn1.Marshal(*pub)
	if err != nil {
		return "", err
	}
	// Onion id is base32(firstHalf(sha1(publicKeyDER)))
	hash := sha1.Sum(der)
	half := hash[:len(hash)/2]
	serviceID := base32.StdEncoding.EncodeToString(half)
	return strings.ToLower(serviceID), nil
}
