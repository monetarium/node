// Copyright (c) 2025 The Monetarium developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"testing"
	"time"

	"github.com/decred/dcrd/chaincfg/v3"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"github.com/decred/dcrd/wire"
)

// TestEmissionAuthorizationBasic tests the basic emission authorization functionality
func TestEmissionAuthorizationBasic(t *testing.T) {
	// Create test parameters
	params := &chaincfg.Params{
		SKAEmissionHeight: 100,
		SKAEmissionAmount: 1000000,
		SKAMaxAmount:      10000000,
	}

	// Generate test keys
	privKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}
	pubKey := privKey.PubKey()

	// Initialize emission keys and nonces
	params.SKAEmissionKeys = map[wire.CoinType]*secp256k1.PublicKey{
		1: pubKey,
	}
	params.SKAEmissionNonces = map[wire.CoinType]uint64{
		1: 0,
	}

	// Create test authorization
	auth := &chaincfg.SKAEmissionAuth{
		EmissionKey: pubKey,
		Signature:   nil, // Will be set after signing
		Nonce:       1,
		CoinType:    1,
		Amount:      1000000,
		Height:      100,
		Timestamp:   time.Now().Unix(),
	}

	// Test addresses and amounts
	addresses := []string{"DsTest1234567890123456789012345678901234567890"}
	amounts := []int64{1000000}

	// Create authorization hash and sign it
	authHash, err := createEmissionAuthHash(auth, addresses, amounts)
	if err != nil {
		t.Fatalf("Failed to create auth hash: %v", err)
	}

	signature := ecdsa.Sign(privKey, authHash[:])
	auth.Signature = signature.Serialize()

	// Test emission authorization validation
	if err := validateEmissionAuthorization(auth, params); err != nil {
		t.Errorf("Valid authorization failed validation: %v", err)
	}

	// Test invalid nonce (replay protection)
	authInvalidNonce := *auth
	authInvalidNonce.Nonce = 0 // Should be 1
	if err := validateEmissionAuthorization(&authInvalidNonce, params); err == nil {
		t.Error("Invalid nonce should have failed validation")
	}

	// Test unauthorized key
	wrongPrivKey, _ := secp256k1.GeneratePrivateKey()
	wrongPubKey := wrongPrivKey.PubKey()
	authWrongKey := *auth
	authWrongKey.EmissionKey = wrongPubKey
	if err := validateEmissionAuthorization(&authWrongKey, params); err == nil {
		t.Error("Wrong key should have failed validation")
	}

	// Test unauthorized coin type
	authWrongCoinType := *auth
	authWrongCoinType.CoinType = 2 // Not configured in params
	if err := validateEmissionAuthorization(&authWrongCoinType, params); err == nil {
		t.Error("Wrong coin type should have failed validation")
	}
}

// TestEmissionAuthorizationScript tests script creation and parsing
func TestEmissionAuthorizationScript(t *testing.T) {
	// Generate test key
	privKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}
	pubKey := privKey.PubKey()

	// Create test authorization
	auth := &chaincfg.SKAEmissionAuth{
		EmissionKey: pubKey,
		Signature:   []byte{1, 2, 3, 4, 5}, // Dummy signature for testing
		Nonce:       12345,
		CoinType:    1,
		Amount:      1000000,
		Height:      100,
		Timestamp:   time.Now().Unix(),
	}

	// Create authorization script
	script, err := createEmissionAuthScript(auth)
	if err != nil {
		t.Fatalf("Failed to create auth script: %v", err)
	}

	// Parse the script back
	parsedAuth, err := extractEmissionAuthorization(script)
	if err != nil {
		t.Fatalf("Failed to extract auth from script: %v", err)
	}

	// Verify extracted data matches original
	if parsedAuth.Nonce != auth.Nonce {
		t.Errorf("Nonce mismatch: got %d, want %d", parsedAuth.Nonce, auth.Nonce)
	}

	if parsedAuth.CoinType != auth.CoinType {
		t.Errorf("CoinType mismatch: got %d, want %d", parsedAuth.CoinType, auth.CoinType)
	}

	if !parsedAuth.EmissionKey.IsEqual(auth.EmissionKey) {
		t.Error("EmissionKey mismatch")
	}

	if len(parsedAuth.Signature) != len(auth.Signature) {
		t.Errorf("Signature length mismatch: got %d, want %d",
			len(parsedAuth.Signature), len(auth.Signature))
	}
}

// TestUnauthorizedEmissionPrevention tests various unauthorized emission scenarios
func TestUnauthorizedEmissionPrevention(t *testing.T) {
	// Create test parameters with no emission keys configured
	params := &chaincfg.Params{
		SKAEmissionHeight: 100,
		SKAEmissionAmount: 1000000,
		SKAMaxAmount:      10000000,
		SKAEmissionKeys:   make(map[wire.CoinType]*secp256k1.PublicKey),
		SKAEmissionNonces: make(map[wire.CoinType]uint64),
	}

	// Test 1: No emission key configured
	privKey, _ := secp256k1.GeneratePrivateKey()
	auth := &chaincfg.SKAEmissionAuth{
		EmissionKey: privKey.PubKey(),
		CoinType:    1,
		Nonce:       1,
	}

	if err := validateEmissionAuthorization(auth, params); err == nil {
		t.Error("Should fail when no emission key is configured")
	}

	// Test 2: Configure a key and test replay protection
	params.SKAEmissionKeys[1] = privKey.PubKey()
	params.SKAEmissionNonces[1] = 5 // Simulate 5 emissions already done

	// Should fail with nonce 5 (replay)
	auth.Nonce = 5
	if err := validateEmissionAuthorization(auth, params); err == nil {
		t.Error("Should fail with replayed nonce")
	}

	// Should fail with nonce 7 (skipped nonce)
	auth.Nonce = 7
	if err := validateEmissionAuthorization(auth, params); err == nil {
		t.Error("Should fail with skipped nonce")
	}

	// Should succeed with nonce 6 (next valid nonce)
	auth.Nonce = 6
	if err := validateEmissionAuthorization(auth, params); err != nil {
		t.Errorf("Should succeed with next valid nonce: %v", err)
	}
}
