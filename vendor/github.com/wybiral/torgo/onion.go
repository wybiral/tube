package torgo

// Onion represents a hidden service.
type Onion struct {
	// Ports maps virtual ports for the hidden service to local addresses.
	Ports map[int]string
	// ServiceID is the unique hidden service address (without ".onion" ending).
	ServiceID string
	// Base64 encoded private key for the hidden service.
	PrivateKey string
	// Type of private key (RSA1024 or ED25519-V3).
	PrivateKeyType string
}
