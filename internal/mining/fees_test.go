// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package mining

import (
	"testing"

	"github.com/decred/dcrd/chaincfg/v3"
	"github.com/decred/dcrd/cointype"
	"github.com/decred/dcrd/txscript/v4/stdaddr"
	"github.com/decred/dcrd/wire"
)

// TestAddFeesToCoinbase tests the addFeesToCoinbase function.
func TestAddFeesToCoinbase(t *testing.T) {
	// Create a mock address for testing - use a simple P2PKH address
	simNetParams := chaincfg.SimNetParams()
	// Create a mock hash for the address
	hash := [20]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a,
		0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14}
	mockAddr, err := stdaddr.NewAddressPubKeyHashEcdsaSecp256k1V0(hash[:], simNetParams)
	if err != nil {
		t.Fatalf("Failed to create mock address: %v", err)
	}

	tests := []struct {
		name            string
		totalFees       wire.FeesByType
		powOutputIdx    int
		expectedVAR     int64
		expectedOutputs int // Expected number of outputs after adding fees
	}{
		{
			name:            "no fees",
			totalFees:       wire.NewFeesByType(),
			powOutputIdx:    1,
			expectedVAR:     0,
			expectedOutputs: 2, // Original outputs only
		},
		{
			name: "VAR fees only",
			totalFees: func() wire.FeesByType {
				fees := wire.NewFeesByType()
				fees.Add(cointype.CoinTypeVAR, 1000)
				return fees
			}(),
			powOutputIdx:    1,
			expectedVAR:     1000,
			expectedOutputs: 2, // No new outputs for VAR
		},
		{
			name: "SKA fees only",
			totalFees: func() wire.FeesByType {
				fees := wire.NewFeesByType()
				fees.Add(cointype.CoinType(1), 500)
				return fees
			}(),
			powOutputIdx:    1,
			expectedVAR:     0,
			expectedOutputs: 3, // One new output for SKA
		},
		{
			name: "mixed fees",
			totalFees: func() wire.FeesByType {
				fees := wire.NewFeesByType()
				fees.Add(cointype.CoinTypeVAR, 1000)
				fees.Add(cointype.CoinType(1), 500)
				fees.Add(cointype.CoinType(2), 300)
				return fees
			}(),
			powOutputIdx:    1,
			expectedVAR:     1000,
			expectedOutputs: 4, // Two new outputs for SKA and coin type 2
		},
		{
			name: "zero fees ignored",
			totalFees: func() wire.FeesByType {
				fees := wire.NewFeesByType()
				fees.Add(cointype.CoinTypeVAR, 1000)
				fees.Add(cointype.CoinType(1), 0) // Should be ignored
				fees.Add(cointype.CoinType(2), 300)
				return fees
			}(),
			powOutputIdx:    1,
			expectedVAR:     1000,
			expectedOutputs: 3, // Only one new output for coin type 2
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create a mock coinbase transaction with initial outputs
			coinbaseTx := &wire.MsgTx{
				TxOut: []*wire.TxOut{
					{Value: 500, CoinType: cointype.CoinTypeVAR},  // Treasury output
					{Value: 1000, CoinType: cointype.CoinTypeVAR}, // PoW output
				},
			}

			initialPowValue := coinbaseTx.TxOut[test.powOutputIdx].Value

			// Call the function under test
			err := addFeesToCoinbase(coinbaseTx, test.totalFees, test.powOutputIdx, mockAddr)
			if err != nil {
				t.Fatalf("addFeesToCoinbase failed: %v", err)
			}

			// Check that the correct number of outputs were created
			if len(coinbaseTx.TxOut) != test.expectedOutputs {
				t.Errorf("Expected %d outputs, got %d", test.expectedOutputs, len(coinbaseTx.TxOut))
			}

			// Check that VAR fees were added to the PoW output
			expectedPowValue := initialPowValue + test.expectedVAR
			if coinbaseTx.TxOut[test.powOutputIdx].Value != expectedPowValue {
				t.Errorf("Expected PoW output value %d, got %d",
					expectedPowValue, coinbaseTx.TxOut[test.powOutputIdx].Value)
			}

			// Check that new outputs have correct coin types and values
			if len(coinbaseTx.TxOut) > 2 {
				for i := 2; i < len(coinbaseTx.TxOut); i++ {
					output := coinbaseTx.TxOut[i]

					// Verify this output corresponds to a fee from totalFees
					expectedAmount := test.totalFees.Get(output.CoinType)
					if expectedAmount <= 0 {
						t.Errorf("Output %d has coin type %d with no corresponding fees", i, output.CoinType)
						continue
					}

					if output.Value != expectedAmount {
						t.Errorf("Output %d: expected value %d, got %d", i, expectedAmount, output.Value)
					}

					// Should not be VAR coin type (VAR goes to PoW output)
					if output.CoinType == cointype.CoinTypeVAR {
						t.Errorf("Output %d should not be VAR coin type", i)
					}
				}
			}
		})
	}
}

// TestFeesByTypeIntegration tests the integration of FeesByType with mining logic.
func TestFeesByTypeIntegration(t *testing.T) {
	// Test that FeesByType methods work correctly in mining context
	fees := wire.NewFeesByType()

	// Add fees for different coin types
	fees.Add(cointype.CoinTypeVAR, 1000)
	fees.Add(cointype.CoinType(1), 500)
	fees.Add(cointype.CoinType(2), 300)

	// Test total calculation
	expectedTotal := int64(1800)
	if fees.Total() != expectedTotal {
		t.Errorf("Expected total fees %d, got %d", expectedTotal, fees.Total())
	}

	// Test coin type retrieval
	if fees.Get(cointype.CoinTypeVAR) != 1000 {
		t.Errorf("Expected VAR fees 1000, got %d", fees.Get(cointype.CoinTypeVAR))
	}

	// Test types enumeration
	types := fees.Types()
	if len(types) != 3 {
		t.Errorf("Expected 3 coin types, got %d", len(types))
	}

	// Test fee scaling (similar to mining logic)
	voters := int64(5)
	ticketsPerBlock := int64(20)

	for coinType := range fees {
		scaledFee := fees[coinType] * voters / ticketsPerBlock
		fees[coinType] = scaledFee
	}

	expectedScaledTotal := expectedTotal * voters / ticketsPerBlock
	if fees.Total() != expectedScaledTotal {
		t.Errorf("Expected scaled total fees %d, got %d", expectedScaledTotal, fees.Total())
	}
}
