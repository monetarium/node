module github.com/monetarium/node/blockchain

go 1.23

toolchain go1.23.4

require (
	github.com/monetarium/node/blockchain/stake v10.0.0
	github.com/monetarium/node/blockchain/standalone v10.0.0
	github.com/monetarium/node/chaincfg/chainhash v10.0.0
	github.com/monetarium/node/chaincfg v10.0.0
	github.com/monetarium/node/cointype v10.0.0
	github.com/monetarium/node/crypto/rand v10.0.0
	github.com/monetarium/node/dcrec v10.0.0
	github.com/monetarium/node/dcrec/secp256k1 v10.0.0
	github.com/monetarium/node/dcrutil v10.0.0
	github.com/monetarium/node/txscript v10.0.0
	github.com/monetarium/node/wire v10.0.0
)

require (
	github.com/agl/ed25519 v0.0.0-20170116200512-5312a6153412 // indirect
	github.com/dchest/siphash v1.2.3 // indirect
	github.com/decred/base58 v1.0.5 // indirect
	github.com/monetarium/node/crypto/blake256 v10.0.0 // indirect
	github.com/monetarium/node/crypto/ripemd160 v10.0.0 // indirect
	github.com/monetarium/node/database v10.0.0 // indirect
	github.com/monetarium/node/dcrec/edwards v10.0.0 // indirect
	github.com/decred/slog v1.2.0 // indirect
	github.com/klauspost/cpuid/v2 v2.0.9 // indirect
	golang.org/x/crypto v0.24.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	lukechampine.com/blake3 v1.3.0 // indirect
)

replace github.com/monetarium/node/blockchain/standalone => ./standalone

replace github.com/monetarium/node/cointype => ../cointype

replace github.com/monetarium/node/database => ../database

replace github.com/monetarium/node/wire => ../wire

replace github.com/monetarium/node/blockchain/stake => ./stake
