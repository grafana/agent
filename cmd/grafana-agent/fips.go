//go:build fips || boringcrypto

package main

// Package fipsonly restricts all TLS configuration to boringcrypto settings.
import _ "crypto/tls/fipsonly"
