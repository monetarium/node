module github.com/decred/dcrd/wire

go 1.23

toolchain go1.23.4

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/decred/dcrd/chaincfg/chainhash v1.0.4
	github.com/decred/dcrd/cointype v1.0.0
	lukechampine.com/blake3 v1.3.0
)

require (
	github.com/decred/dcrd/crypto/blake256 v1.0.1 // indirect
	github.com/klauspost/cpuid/v2 v2.0.9 // indirect
)

replace github.com/decred/dcrd/dcrutil/v4 => ../dcrutil

replace github.com/decred/dcrd/cointype => ../cointype
