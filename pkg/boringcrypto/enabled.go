//go:build fips || boringcrypto

package boringcrypto

// Package fipsonly restricts all TLS configuration to boringcrypto settings.
import _ "crypto/tls/fipsonly"

const Enabled = true
