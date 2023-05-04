//go:build linux
// +build linux

package ethtool

import (
	"fmt"

	"github.com/mdlayher/netlink"
	"github.com/mdlayher/netlink/nlenc"
	"golang.org/x/sys/unix"
)

// A bitset is a compact bitset used by ethtool netlink.
type bitset []uint32

// newBitset creates a bitset from a netlink attribute decoder by parsing the
// various fields and applying a mask to the set if needed.
func newBitset(ad *netlink.AttributeDecoder) (bitset, error) {
	// Bitsets are represented as a slice of contiguous uint32 which each
	// contain bits. By default, the mask bitset is applied to values unless
	// we explicitly find the NOMASK flag.
	var (
		values, mask bitset
		doMask       = true
	)

	for ad.Next() {
		switch ad.Type() {
		case unix.ETHTOOL_A_BITSET_NOMASK:
			doMask = false
		case unix.ETHTOOL_A_BITSET_SIZE:
			// Convert number of bits to number of bytes, rounded up to the
			// nearest 32 bits for a uint32 boundary.
			n := (ad.Uint32() + 31) / 32
			values = make(bitset, n)
			if doMask {
				mask = make(bitset, n)
			}
		case unix.ETHTOOL_A_BITSET_VALUE:
			ad.Do(values.decode)
		case unix.ETHTOOL_A_BITSET_MASK:
			ad.Do(mask.decode)
		}
	}

	// Do a quick check for errors before making use of the bitsets. Normally
	// this will be called in a nested attribute decoder context and we could
	// skip this, but we don't want to return an invalid bitset.
	if err := ad.Err(); err != nil {
		return nil, err
	}

	// Mask by default unless the caller told us not to.
	if doMask {
		for i := 0; i < len(values); i++ {
			values[i] &= mask[i]
		}
	}

	return values, nil
}

// decode returns a function which parses a compact bitset into a preallocated
// bitset. The bitset must be preallocated with the appropriate length and
// capacity for the length of the input data.
func (bs *bitset) decode(b []byte) error {
	if len(b)/4 != len(*bs) {
		return fmt.Errorf("ethtool: cannot store %d bytes in bitset with length %d",
			len(b), len(*bs))
	}

	for i := 0; i < len(*bs); i++ {
		(*bs)[i] = nlenc.Uint32(b[i*4 : (i*4)+4])
	}

	return nil
}

// test is like the ethnl_bitmap32_test_bit() function in the Linux kernel: it
// reports whether the bit with the specified index is set in the bitset.
func (bs *bitset) test(idx int) bool {
	return (*bs)[idx/32]&(1<<(idx%32)) != 0
}
