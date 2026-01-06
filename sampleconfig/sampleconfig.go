// Copyright (c) 2017-2022 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package sampleconfig

import (
	_ "embed"
)

// sampleMonetariumConf is a string containing the commented example config for monetarium.
//
//go:embed sample-monetarium.conf
var sampleMonetariumConf string

// sampleDcrctlConf is a string containing the commented example config for
// dcrctl.
//
//go:embed sample-dcrctl.conf
var sampleDcrctlConf string

// Dcrd returns a string containing the commented example config for monetarium.
func Dcrd() string {
	return sampleMonetariumConf
}

// FileContents returns a string containing the commented example config for
// dcrd.
//
// Deprecated: Use the [Dcrd] function instead.
func FileContents() string {
	return Dcrd()
}

// Dcrctl returns a string containing the commented example config for dcrctl.
func Dcrctl() string {
	return sampleDcrctlConf
}

// DcrctlSampleConfig is a string containing the commented example config for
// dcrctl.
//
// Deprecated: Use the [Dcrctl] function instead.
var DcrctlSampleConfig = Dcrctl()
