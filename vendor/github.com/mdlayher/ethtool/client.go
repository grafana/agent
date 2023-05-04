package ethtool

import (
	"fmt"
)

//go:generate stringer -type=Duplex,Port -output=string.go
//go:generate go run mklinkmodes.go

var (
	_ error = &Error{}
	// Ensure compatibility with Go 1.13+ errors package.
	_ interface{ Unwrap() error } = &Error{}
)

// An Error is an error value produced by the kernel due to a bad ethtool
// netlink request. Typically the Err will be of type *netlink.OpError.
type Error struct {
	Message string
	Err     error
}

// Error implements error.
func (e *Error) Error() string {
	// This typically wraps a *netlink.OpError which will contain the error
	// string anyway, so just return the inner error's string.
	return e.Err.Error()
}

// Unwrap unwraps the internal Err field for use with errors.Unwrap.
func (e *Error) Unwrap() error { return e.Err }

// A Client can manipulate the ethtool netlink interface.
type Client struct {
	// The operating system-specific client.
	c *client
}

// New creates a Client which can issue ethtool commands.
func New() (*Client, error) {
	c, err := newClient()
	if err != nil {
		return nil, err
	}

	return &Client{c: c}, nil
}

// An Interface is an ethtool netlink Ethernet interface. Interfaces are used to
// identify an Ethernet interface being queried by its index and/or name.
type Interface struct {
	// Callers may choose to set either Index, Name, or both fields. Note that
	// if both are set, the kernel will verify that both Index and Name are
	// associated with the same interface. If they are not, an error will be
	// returned.
	Index int
	Name  string
}

// LinkInfo contains link settings for an Ethernet interface.
type LinkInfo struct {
	Interface Interface
	Port      Port
}

// A Port is the port type for a LinkInfo structure.
type Port int

// Possible Port type values.
const (
	TwistedPair  Port = 0x00
	AUI          Port = 0x01
	MII          Port = 0x02
	Fibre        Port = 0x03
	BNC          Port = 0x04
	DirectAttach Port = 0x05
	None         Port = 0xef
	Other        Port = 0xff
)

// LinkInfos fetches LinkInfo structures for each ethtool-supported interface
// on this system.
func (c *Client) LinkInfos() ([]*LinkInfo, error) {
	return c.c.LinkInfos()
}

// LinkInfo fetches LinkInfo for the specified Interface.
//
// If the requested device does not exist or is not supported by the ethtool
// interface, an error compatible with errors.Is(err, os.ErrNotExist) will be
// returned.
func (c *Client) LinkInfo(ifi Interface) (*LinkInfo, error) {
	return c.c.LinkInfo(ifi)
}

// LinkMode contains link mode information for an Ethernet interface.
type LinkMode struct {
	Interface     Interface
	SpeedMegabits int
	Ours, Peer    []AdvertisedLinkMode
	Duplex        Duplex
}

// A Duplex is the link duplex type for a LinkMode structure.
type Duplex int

// Possible Duplex type values.
const (
	Half    Duplex = 0x00
	Full    Duplex = 0x01
	Unknown Duplex = 0xff
)

// An AdvertisedLinkMode is a link mode that an interface advertises it is
// capable of using.
type AdvertisedLinkMode struct {
	Index int
	Name  string
}

// LinkModes fetches LinkMode structures for each ethtool-supported interface
// on this system.
func (c *Client) LinkModes() ([]*LinkMode, error) {
	return c.c.LinkModes()
}

// LinkMode fetches LinkMode data for the specified Interface.
//
// If the requested device does not exist or is not supported by the ethtool
// interface, an error compatible with errors.Is(err, os.ErrNotExist) will be
// returned.
func (c *Client) LinkMode(ifi Interface) (*LinkMode, error) {
	return c.c.LinkMode(ifi)
}

// LinkState contains link state information for an Ethernet interface.
type LinkState struct {
	Interface Interface
	Link      bool
}

// LinkStates fetches LinkState structures for each ethtool-supported interface
// on this system.
func (c *Client) LinkStates() ([]*LinkState, error) {
	return c.c.LinkStates()
}

// LinkState fetches LinkState data for the specified Interface.
//
// If the requested device does not exist or is not supported by the ethtool
// interface, an error compatible with errors.Is(err, os.ErrNotExist) will be
// returned.
func (c *Client) LinkState(ifi Interface) (*LinkState, error) {
	return c.c.LinkState(ifi)
}

// FEC fetches the forward error correction (FEC) setting for the specified
// Interface.
func (c *Client) FEC(ifi Interface) (*FEC, error) {
	return c.c.FEC(ifi)
}

// SetFEC sets the forward error correction (FEC) parameters for the Interface
// in fec.
//
// Setting FEC parameters requires elevated privileges and if the caller
// does not have permission, an error compatible with errors.Is(err,
// os.ErrPermission) will be returned.
//
// If the requested device does not exist or is not supported by the ethtool
// interface, an error compatible with errors.Is(err, os.ErrNotExist) will be
// returned.
func (c *Client) SetFEC(fec FEC) error {
	return c.c.SetFEC(fec)
}

// A FEC contains the forward error correction (FEC) parameters for an
// interface.
type FEC struct {
	Interface Interface
	Modes     FECModes
	Active    FECMode
	Auto      bool
}

// A FECMode is a FEC mode bit value (single element bitmask) specifying the
// active mode of an interface.
type FECMode int

// A FECModes is a FEC mode bitmask of mode(s) supported by an interface.
type FECModes FECMode

// A WakeOnLAN contains the Wake-on-LAN parameters for an interface.
type WakeOnLAN struct {
	Interface Interface
	Modes     WOLMode
}

// A WOLMode is a Wake-on-LAN mode bitmask of mode(s) supported by an interface.
type WOLMode int

// Possible Wake-on-LAN mode bit flags.
const (
	PHY         WOLMode = 1 << 0
	Unicast     WOLMode = 1 << 1
	Multicast   WOLMode = 1 << 2
	Broadcast   WOLMode = 1 << 3
	ARP         WOLMode = 1 << 4
	Magic       WOLMode = 1 << 5
	MagicSecure WOLMode = 1 << 6
	Filter      WOLMode = 1 << 7
)

// String returns the string representation of a WOLMode bitmask.
func (m WOLMode) String() string {
	names := []string{
		"PHY",
		"Unicast",
		"Multicast",
		"Broadcast",
		"ARP",
		"Magic",
		"MagicSecure",
		"Filter",
	}

	var s string
	left := uint(m)
	for i, name := range names {
		if m&(1<<uint(i)) != 0 {
			if s != "" {
				s += "|"
			}

			s += name

			left ^= (1 << uint(i))
		}
	}

	if s == "" && left == 0 {
		s = "0"
	}

	if left > 0 {
		if s != "" {
			s += "|"
		}
		s += fmt.Sprintf("%#x", left)
	}

	return s
}

// WakeOnLANs fetches WakeOnLAN information for each ethtool-supported interface
// on this system.
func (c *Client) WakeOnLANs() ([]*WakeOnLAN, error) {
	return c.c.WakeOnLANs()
}

// WakeOnLAN fetches WakeOnLAN parameters for the specified Interface.
//
// Fetching Wake-on-LAN information requires elevated privileges and if the
// caller does not have permission, an error compatible with errors.Is(err,
// os.ErrPermission) will be returned.
//
// If the requested device does not exist or is not supported by the ethtool
// interface, an error compatible with errors.Is(err, os.ErrNotExist) will be
// returned.
func (c *Client) WakeOnLAN(ifi Interface) (*WakeOnLAN, error) {
	return c.c.WakeOnLAN(ifi)
}

// SetWakeOnLAN sets the WakeOnLAN parameters for the Interface in wol.
//
// Setting Wake-on-LAN parameters requires elevated privileges and if the caller
// does not have permission, an error compatible with errors.Is(err,
// os.ErrPermission) will be returned.
//
// If the requested device does not exist or is not supported by the ethtool
// interface, an error compatible with errors.Is(err, os.ErrNotExist) will be
// returned.
func (c *Client) SetWakeOnLAN(wol WakeOnLAN) error {
	return c.c.SetWakeOnLAN(wol)
}

// Close cleans up the Client's resources.
func (c *Client) Close() error { return c.c.Close() }
