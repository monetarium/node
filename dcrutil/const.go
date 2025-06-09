// Copyright (c) 2013-2014 The btcsuite developers
// Copyright (c) 2015-2024 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package dcrutil

const (
	// AtomsPerCent is the number of atomic units in one coin cent.
	AtomsPerCent = 1e6

	// AtomsPerCoin is the number of atomic units in one coin.
	// This constant maintains backward compatibility and refers to VAR.
	AtomsPerCoin = 1e8

	// MaxAmount is the maximum transaction amount allowed in atoms.
	// This constant maintains backward compatibility and refers to VAR.
	MaxAmount = 21e6 * AtomsPerCoin

	// SatoshiPerVAR is an alias for AtomsPerVAR for compatibility.
	SatoshiPerVAR = AtomsPerVAR

	// SatoshiPerSKA is an alias for AtomsPerSKA for compatibility.
	SatoshiPerSKA = AtomsPerSKA
)
