//go:build !(fips || boringcrypto || cngcrypto)

package boringcrypto

const Enabled = false
