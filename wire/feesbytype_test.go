// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"testing"
)

// TestFeesByType tests the basic functionality of FeesByType.
func TestFeesByType(t *testing.T) {
	// Test NewFeesByType
	fees := NewFeesByType()
	if fees == nil {
		t.Fatal("NewFeesByType returned nil")
	}
	if len(fees) != 0 {
		t.Errorf("Expected empty fees map, got length %d", len(fees))
	}

	// Test Add and Get
	fees.Add(CoinTypeVAR, 1000)
	fees.Add(CoinTypeSKA, 500)
	fees.Add(CoinType(2), 300)

	if got := fees.Get(CoinTypeVAR); got != 1000 {
		t.Errorf("Expected VAR fees 1000, got %d", got)
	}
	if got := fees.Get(CoinTypeSKA); got != 500 {
		t.Errorf("Expected SKA fees 500, got %d", got)
	}
	if got := fees.Get(CoinType(2)); got != 300 {
		t.Errorf("Expected coin type 2 fees 300, got %d", got)
	}

	// Test Add to existing
	fees.Add(CoinTypeVAR, 200)
	if got := fees.Get(CoinTypeVAR); got != 1200 {
		t.Errorf("Expected VAR fees 1200 after adding 200, got %d", got)
	}

	// Test Get for non-existent coin type
	if got := fees.Get(CoinType(99)); got != 0 {
		t.Errorf("Expected 0 for non-existent coin type, got %d", got)
	}
}

// TestFeesByTypeTotal tests the Total method.
func TestFeesByTypeTotal(t *testing.T) {
	fees := NewFeesByType()

	// Test empty total
	if got := fees.Total(); got != 0 {
		t.Errorf("Expected total 0 for empty fees, got %d", got)
	}

	// Add fees and test total
	fees.Add(CoinTypeVAR, 1000)
	fees.Add(CoinTypeSKA, 500)
	fees.Add(CoinType(2), 300)

	expected := int64(1800)
	if got := fees.Total(); got != expected {
		t.Errorf("Expected total %d, got %d", expected, got)
	}
}

// TestFeesByTypeTypes tests the Types method.
func TestFeesByTypeTypes(t *testing.T) {
	fees := NewFeesByType()

	// Test empty types
	types := fees.Types()
	if len(types) != 0 {
		t.Errorf("Expected no types for empty fees, got %d", len(types))
	}

	// Add fees and test types
	fees.Add(CoinTypeVAR, 1000)
	fees.Add(CoinTypeSKA, 500)
	fees.Add(CoinType(2), 0) // Zero amount should not be included
	fees.Add(CoinType(3), 300)

	types = fees.Types()
	expectedCount := 3 // VAR, SKA, and coin type 3 (not 2 since it's zero)
	if len(types) != expectedCount {
		t.Errorf("Expected %d types, got %d", expectedCount, len(types))
	}

	// Check that all expected types are present
	typeSet := make(map[CoinType]bool)
	for _, coinType := range types {
		typeSet[coinType] = true
	}

	expectedTypes := []CoinType{CoinTypeVAR, CoinTypeSKA, CoinType(3)}
	for _, expected := range expectedTypes {
		if !typeSet[expected] {
			t.Errorf("Expected coin type %d in types, but not found", expected)
		}
	}

	// Check that zero-amount type is not included
	if typeSet[CoinType(2)] {
		t.Error("Expected coin type 2 (zero amount) not to be included in types")
	}
}

// TestFeesByTypeMerge tests the Merge method.
func TestFeesByTypeMerge(t *testing.T) {
	fees1 := NewFeesByType()
	fees1.Add(CoinTypeVAR, 1000)
	fees1.Add(CoinTypeSKA, 500)

	fees2 := NewFeesByType()
	fees2.Add(CoinTypeVAR, 200) // Should add to existing
	fees2.Add(CoinType(2), 300) // New coin type

	fees1.Merge(fees2)

	// Check merged results
	if got := fees1.Get(CoinTypeVAR); got != 1200 {
		t.Errorf("Expected merged VAR fees 1200, got %d", got)
	}
	if got := fees1.Get(CoinTypeSKA); got != 500 {
		t.Errorf("Expected SKA fees unchanged at 500, got %d", got)
	}
	if got := fees1.Get(CoinType(2)); got != 300 {
		t.Errorf("Expected new coin type 2 fees 300, got %d", got)
	}

	// Original fees2 should be unchanged
	if got := fees2.Get(CoinTypeVAR); got != 200 {
		t.Errorf("Expected fees2 VAR unchanged at 200, got %d", got)
	}
}

// TestGetPrimaryCoinType tests the GetPrimaryCoinType function.
func TestGetPrimaryCoinType(t *testing.T) {
	tests := []struct {
		name     string
		outputs  []*TxOut
		expected CoinType
	}{
		{
			name:     "empty transaction",
			outputs:  []*TxOut{},
			expected: CoinTypeVAR,
		},
		{
			name: "VAR only transaction",
			outputs: []*TxOut{
				{Value: 1000, CoinType: CoinTypeVAR},
				{Value: 500, CoinType: CoinTypeVAR},
			},
			expected: CoinTypeVAR,
		},
		{
			name: "SKA transaction",
			outputs: []*TxOut{
				{Value: 1000, CoinType: CoinTypeSKA},
				{Value: 500, CoinType: CoinTypeSKA},
			},
			expected: CoinTypeSKA,
		},
		{
			name: "mixed transaction - first non-VAR wins",
			outputs: []*TxOut{
				{Value: 1000, CoinType: CoinTypeVAR},
				{Value: 500, CoinType: CoinType(2)},
				{Value: 300, CoinType: CoinTypeSKA},
			},
			expected: CoinType(2),
		},
		{
			name: "SKA-3 transaction",
			outputs: []*TxOut{
				{Value: 1000, CoinType: CoinType(3)},
			},
			expected: CoinType(3),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tx := &MsgTx{
				TxOut: test.outputs,
			}

			result := GetPrimaryCoinType(tx)
			if result != test.expected {
				t.Errorf("Expected coin type %d, got %d", test.expected, result)
			}
		})
	}
}

// TestFeesByTypeEdgeCases tests edge cases and error conditions.
func TestFeesByTypeEdgeCases(t *testing.T) {
	fees := NewFeesByType()

	// Test adding negative fees (should still work)
	fees.Add(CoinTypeVAR, -100)
	if got := fees.Get(CoinTypeVAR); got != -100 {
		t.Errorf("Expected negative fees -100, got %d", got)
	}

	// Test large coin type values
	largeCoinType := CoinType(255) // Maximum coin type
	fees.Add(largeCoinType, 1000)
	if got := fees.Get(largeCoinType); got != 1000 {
		t.Errorf("Expected fees for large coin type, got %d", got)
	}

	// Test Types() with negative values
	types := fees.Types()
	expectedCount := 1 // Only the large coin type has positive fees
	if len(types) != expectedCount {
		t.Errorf("Expected %d types with positive fees, got %d", expectedCount, len(types))
	}

	// Test Total with negative values
	expectedTotal := int64(900) // -100 + 1000
	if got := fees.Total(); got != expectedTotal {
		t.Errorf("Expected total %d, got %d", expectedTotal, got)
	}
}
