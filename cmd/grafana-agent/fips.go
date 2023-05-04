//go:build fips || boringcrypto

package main

import _ "crypto/tls/fipsonly"
