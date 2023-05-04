//go:build !linux
// +build !linux

package wifi

import (
	"fmt"
	"runtime"
)

// errUnimplemented is returned by all functions on platforms that
// do not have package wifi implemented.
var errUnimplemented = fmt.Errorf("wifi: not implemented on %s", runtime.GOOS)

// A conn is the no-op implementation of a netlink sockets connection.
type client struct{}

func newClient() (*client, error) { return nil, errUnimplemented }

func (*client) Close() error                                     { return errUnimplemented }
func (*client) Interfaces() ([]*Interface, error)                { return nil, errUnimplemented }
func (*client) BSS(_ *Interface) (*BSS, error)                   { return nil, errUnimplemented }
func (*client) StationInfo(_ *Interface) ([]*StationInfo, error) { return nil, errUnimplemented }
func (*client) Connect(_ *Interface, _ string) error             { return errUnimplemented }
func (*client) Disconnect(_ *Interface) error                    { return errUnimplemented }
func (*client) ConnectWPAPSK(_ *Interface, _, _ string) error    { return errUnimplemented }
