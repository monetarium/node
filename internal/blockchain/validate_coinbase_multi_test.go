// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"errors"
	"testing"

	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/cointype"
	"github.com/decred/dcrd/wire"
)

// TestValidateCoinbaseMultiOutput tests the validateCoinbaseMultiOutput function
// with various valid and invalid coinbase transaction structures.
func TestValidateCoinbaseMultiOutput(t *testing.T) {
	t.Log("=== Testing Detailed Coinbase Multi-Output Validation ===")

	// Helper to create a basic coinbase transaction
	createCoinbase := func(outputs []*wire.TxOut) *wire.MsgTx {
		coinbase := &wire.MsgTx{
			Version: 1,
			TxIn: []*wire.TxIn{{
				PreviousOutPoint: wire.OutPoint{
					Hash:  chainhash.Hash{},
					Index: 0xffffffff,
				},
				SignatureScript: []byte{0x00, 0x00}, // Minimal coinbase script
				Sequence:        0xffffffff,
				ValueIn:         1000000, // 1 DCR subsidy
			}},
			TxOut:    outputs,
			LockTime: 0,
		}
		return coinbase
	}

	// Standard payment script for testing
	payScript := []byte{0x76, 0xa9, 0x14} // OP_DUP OP_HASH160 <20 bytes> OP_EQUALVERIFY OP_CHECKSIG (simplified)
	for i := 0; i < 20; i++ {
		payScript = append(payScript, byte(i))
	}
	payScript = append(payScript, 0x88, 0xac)

	t.Run("Valid single output coinbase (VAR only)", func(t *testing.T) {
		// Test case: Only VAR subsidy, no fees
		outputs := []*wire.TxOut{
			{
				Value:    1000000, // 1 DCR subsidy
				CoinType: cointype.CoinTypeVAR,
				Version:  0,
				PkScript: payScript,
			},
		}

		coinbase := createCoinbase(outputs)
		expectedFees := wire.NewFeesByType()
		subsidyAmount := int64(1000000)

		err := validateCoinbaseMultiOutput(coinbase, expectedFees, subsidyAmount)
		if err != nil {
			t.Errorf("Valid single output coinbase failed validation: %v", err)
		}

		t.Log("✅ Single VAR output coinbase validation passed")
	})

	t.Run("Valid multi-output coinbase with fees", func(t *testing.T) {
		// Test case: VAR subsidy + VAR fees + SKA fees
		outputs := []*wire.TxOut{
			{
				Value:    1010000, // 1 DCR subsidy + 0.1 DCR VAR fees
				CoinType: cointype.CoinTypeVAR,
				Version:  0,
				PkScript: payScript,
			},
			{
				Value:    5000, // SKA-1 fees
				CoinType: cointype.CoinType(1),
				Version:  0,
				PkScript: payScript,
			},
			{
				Value:    3000, // SKA-2 fees
				CoinType: cointype.CoinType(2),
				Version:  0,
				PkScript: payScript,
			},
		}

		coinbase := createCoinbase(outputs)
		expectedFees := wire.NewFeesByType()
		expectedFees.Add(cointype.CoinTypeVAR, 10000) // 0.1 DCR VAR fees
		expectedFees.Add(cointype.CoinType(1), 5000)  // SKA-1 fees
		expectedFees.Add(cointype.CoinType(2), 3000)  // SKA-2 fees
		subsidyAmount := int64(1000000)

		err := validateCoinbaseMultiOutput(coinbase, expectedFees, subsidyAmount)
		if err != nil {
			t.Errorf("Valid multi-output coinbase failed validation: %v", err)
		}

		t.Log("✅ Multi-output coinbase with mixed fees validation passed")
	})

	t.Run("Invalid: No outputs", func(t *testing.T) {
		coinbase := createCoinbase([]*wire.TxOut{})
		expectedFees := wire.NewFeesByType()
		subsidyAmount := int64(1000000)

		err := validateCoinbaseMultiOutput(coinbase, expectedFees, subsidyAmount)
		if err == nil {
			t.Error("Expected error for coinbase with no outputs")
		}

		if !errors.Is(err, ErrBadCoinbaseOutputStructure) {
			t.Errorf("Expected ErrBadCoinbaseOutputStructure, got: %v", err)
		}

		t.Log("✅ No outputs validation correctly failed")
	})

	t.Run("Invalid: Missing VAR subsidy output", func(t *testing.T) {
		// Test case: Only SKA outputs, no VAR subsidy
		outputs := []*wire.TxOut{
			{
				Value:    5000,
				CoinType: cointype.CoinType(1),
				Version:  0,
				PkScript: payScript,
			},
		}

		coinbase := createCoinbase(outputs)
		expectedFees := wire.NewFeesByType()
		expectedFees.Add(cointype.CoinType(1), 5000)
		subsidyAmount := int64(1000000)

		err := validateCoinbaseMultiOutput(coinbase, expectedFees, subsidyAmount)
		if err == nil {
			t.Error("Expected error for coinbase missing VAR subsidy output")
		}

		if !errors.Is(err, ErrBadCoinbaseOutputStructure) {
			t.Errorf("Expected ErrBadCoinbaseOutputStructure, got: %v", err)
		}

		t.Log("✅ Missing VAR subsidy validation correctly failed")
	})

	t.Run("Invalid: Duplicate coin types", func(t *testing.T) {
		// Test case: Two outputs with same coin type
		outputs := []*wire.TxOut{
			{
				Value:    1000000,
				CoinType: cointype.CoinTypeVAR,
				Version:  0,
				PkScript: payScript,
			},
			{
				Value:    5000,
				CoinType: cointype.CoinType(1),
				Version:  0,
				PkScript: payScript,
			},
			{
				Value:    3000,
				CoinType: cointype.CoinType(1), // Duplicate!
				Version:  0,
				PkScript: payScript,
			},
		}

		coinbase := createCoinbase(outputs)
		expectedFees := wire.NewFeesByType()
		expectedFees.Add(cointype.CoinType(1), 8000)
		subsidyAmount := int64(1000000)

		err := validateCoinbaseMultiOutput(coinbase, expectedFees, subsidyAmount)
		if err == nil {
			t.Error("Expected error for duplicate coin types")
		}

		if !errors.Is(err, ErrBadCoinbaseOutputStructure) {
			t.Errorf("Expected ErrBadCoinbaseOutputStructure, got: %v", err)
		}

		t.Log("✅ Duplicate coin types validation correctly failed")
	})

	t.Run("Invalid: Wrong subsidy amount", func(t *testing.T) {
		// Test case: Subsidy output has wrong amount
		outputs := []*wire.TxOut{
			{
				Value:    900000, // Wrong amount (should be 1000000 + VAR fees)
				CoinType: cointype.CoinTypeVAR,
				Version:  0,
				PkScript: payScript,
			},
		}

		coinbase := createCoinbase(outputs)
		expectedFees := wire.NewFeesByType()
		expectedFees.Add(cointype.CoinTypeVAR, 10000)
		subsidyAmount := int64(1000000)

		err := validateCoinbaseMultiOutput(coinbase, expectedFees, subsidyAmount)
		if err == nil {
			t.Error("Expected error for wrong subsidy amount")
		}

		if !errors.Is(err, ErrBadCoinbaseFeeDistribution) {
			t.Errorf("Expected ErrBadCoinbaseFeeDistribution, got: %v", err)
		}

		t.Log("✅ Wrong subsidy amount validation correctly failed")
	})

	t.Run("Invalid: Wrong fee amount", func(t *testing.T) {
		// Test case: Fee output has wrong amount
		outputs := []*wire.TxOut{
			{
				Value:    1000000, // Correct subsidy (no VAR fees)
				CoinType: cointype.CoinTypeVAR,
				Version:  0,
				PkScript: payScript,
			},
			{
				Value:    3000, // Wrong amount (expected 5000)
				CoinType: cointype.CoinType(1),
				Version:  0,
				PkScript: payScript,
			},
		}

		coinbase := createCoinbase(outputs)
		expectedFees := wire.NewFeesByType()
		expectedFees.Add(cointype.CoinType(1), 5000) // Expected 5000, got 3000
		subsidyAmount := int64(1000000)

		err := validateCoinbaseMultiOutput(coinbase, expectedFees, subsidyAmount)
		if err == nil {
			t.Error("Expected error for wrong fee amount")
		}

		if !errors.Is(err, ErrBadCoinbaseFeeDistribution) {
			t.Errorf("Expected ErrBadCoinbaseFeeDistribution, got: %v", err)
		}

		t.Log("✅ Wrong fee amount validation correctly failed")
	})

	t.Run("Invalid: Unexpected fee output", func(t *testing.T) {
		// Test case: Fee output for coin type that didn't pay fees
		outputs := []*wire.TxOut{
			{
				Value:    1000000,
				CoinType: cointype.CoinTypeVAR,
				Version:  0,
				PkScript: payScript,
			},
			{
				Value:    5000, // Unexpected output
				CoinType: cointype.CoinType(1),
				Version:  0,
				PkScript: payScript,
			},
		}

		coinbase := createCoinbase(outputs)
		expectedFees := wire.NewFeesByType() // No fees expected for any coin type
		subsidyAmount := int64(1000000)

		err := validateCoinbaseMultiOutput(coinbase, expectedFees, subsidyAmount)
		if err == nil {
			t.Error("Expected error for unexpected fee output")
		}

		if !errors.Is(err, ErrBadCoinbaseFeeDistribution) {
			t.Errorf("Expected ErrBadCoinbaseFeeDistribution, got: %v", err)
		}

		t.Log("✅ Unexpected fee output validation correctly failed")
	})

	t.Run("Invalid: Different payment scripts", func(t *testing.T) {
		// Test case: Outputs have different payment scripts
		altPayScript := make([]byte, len(payScript))
		copy(altPayScript, payScript)
		altPayScript[5] = 0xFF // Change one byte

		outputs := []*wire.TxOut{
			{
				Value:    1005000,
				CoinType: cointype.CoinTypeVAR,
				Version:  0,
				PkScript: payScript,
			},
			{
				Value:    5000,
				CoinType: cointype.CoinType(1),
				Version:  0,
				PkScript: altPayScript, // Different script!
			},
		}

		coinbase := createCoinbase(outputs)
		expectedFees := wire.NewFeesByType()
		expectedFees.Add(cointype.CoinTypeVAR, 5000)
		expectedFees.Add(cointype.CoinType(1), 5000)
		subsidyAmount := int64(1000000)

		err := validateCoinbaseMultiOutput(coinbase, expectedFees, subsidyAmount)
		if err == nil {
			t.Error("Expected error for different payment scripts")
		}

		if !errors.Is(err, ErrBadCoinbaseMultiOutput) {
			t.Errorf("Expected ErrBadCoinbaseMultiOutput, got: %v", err)
		}

		t.Log("✅ Different payment scripts validation correctly failed")
	})

	// Note: Invalid coin type test case removed because CoinType is uint8,
	// so values >255 are not representable and the Go compiler prevents this.

	t.Run("Valid: Maximum coin types scenario", func(t *testing.T) {
		// Test case: Multiple valid coin types (stress test)
		outputs := make([]*wire.TxOut, 0)

		// Add VAR subsidy + fees
		outputs = append(outputs, &wire.TxOut{
			Value:    1050000, // 1 DCR subsidy + 0.5 DCR VAR fees
			CoinType: cointype.CoinTypeVAR,
			Version:  0,
			PkScript: payScript,
		})

		// Add multiple SKA coin types
		expectedFees := wire.NewFeesByType()
		expectedFees.Add(cointype.CoinTypeVAR, 50000)

		for i := 1; i <= 10; i++ {
			feeAmount := int64(i * 1000) // Varying fee amounts
			outputs = append(outputs, &wire.TxOut{
				Value:    feeAmount,
				CoinType: cointype.CoinType(i),
				Version:  0,
				PkScript: payScript,
			})
			expectedFees.Add(cointype.CoinType(i), feeAmount)
		}

		coinbase := createCoinbase(outputs)
		subsidyAmount := int64(1000000)

		err := validateCoinbaseMultiOutput(coinbase, expectedFees, subsidyAmount)
		if err != nil {
			t.Errorf("Valid maximum coin types scenario failed: %v", err)
		}

		t.Log("✅ Maximum coin types scenario validation passed")
	})
}

// TestValidateCoinbaseMultiOutputEdgeCases tests edge cases and boundary conditions
func TestValidateCoinbaseMultiOutputEdgeCases(t *testing.T) {
	t.Log("=== Testing Edge Cases for Coinbase Multi-Output Validation ===")

	// Standard payment script for testing
	payScript := []byte{0x76, 0xa9, 0x14}
	for i := 0; i < 20; i++ {
		payScript = append(payScript, byte(i))
	}
	payScript = append(payScript, 0x88, 0xac)

	t.Run("Edge case: Zero fee amounts", func(t *testing.T) {
		// Test that zero fee amounts are handled correctly
		outputs := []*wire.TxOut{
			{
				Value:    1000000, // Just subsidy, no fees
				CoinType: cointype.CoinTypeVAR,
				Version:  0,
				PkScript: payScript,
			},
		}

		coinbase := &wire.MsgTx{
			Version: 1,
			TxIn: []*wire.TxIn{{
				PreviousOutPoint: wire.OutPoint{
					Hash:  chainhash.Hash{},
					Index: 0xffffffff,
				},
				SignatureScript: []byte{0x00, 0x00},
				Sequence:        0xffffffff,
				ValueIn:         1000000,
			}},
			TxOut:    outputs,
			LockTime: 0,
		}

		expectedFees := wire.NewFeesByType()
		subsidyAmount := int64(1000000)

		err := validateCoinbaseMultiOutput(coinbase, expectedFees, subsidyAmount)
		if err != nil {
			t.Errorf("Zero fee amounts should be valid: %v", err)
		}

		t.Log("✅ Zero fee amounts handled correctly")
	})

	t.Run("Edge case: Maximum coin type (255)", func(t *testing.T) {
		// Test the boundary of valid coin types
		outputs := []*wire.TxOut{
			{
				Value:    1000000,
				CoinType: cointype.CoinTypeVAR,
				Version:  0,
				PkScript: payScript,
			},
			{
				Value:    1000,
				CoinType: cointype.CoinType(255), // Maximum valid coin type
				Version:  0,
				PkScript: payScript,
			},
		}

		coinbase := &wire.MsgTx{
			Version: 1,
			TxIn: []*wire.TxIn{{
				PreviousOutPoint: wire.OutPoint{
					Hash:  chainhash.Hash{},
					Index: 0xffffffff,
				},
				SignatureScript: []byte{0x00, 0x00},
				Sequence:        0xffffffff,
				ValueIn:         1000000,
			}},
			TxOut:    outputs,
			LockTime: 0,
		}

		expectedFees := wire.NewFeesByType()
		expectedFees.Add(cointype.CoinType(255), 1000)
		subsidyAmount := int64(1000000)

		err := validateCoinbaseMultiOutput(coinbase, expectedFees, subsidyAmount)
		if err != nil {
			t.Errorf("Maximum coin type 255 should be valid: %v", err)
		}

		t.Log("✅ Maximum coin type (255) handled correctly")
	})
}
