package symtab

import (
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

// ProcMapPermissions contains permission settings read from `/proc/[pid]/maps`.
type ProcMapPermissions struct {
	// mapping has the [R]ead flag set
	Read bool
	// mapping has the [W]rite flag set
	Write bool
	// mapping has the [X]ecutable flag set
	Execute bool
	// mapping has the [S]hared flag set
	Shared bool
	// mapping is marked as [P]rivate (copy on write)
	Private bool
}

// ProcMap contains the process memory-mappings of the process
// read from `/proc/[pid]/maps`.
type ProcMap struct {
	// The start address of current mapping.
	StartAddr uint64
	// The end address of the current mapping
	EndAddr uint64
	// The permissions for this mapping
	Perms *ProcMapPermissions
	// The current offset into the file/fd (e.g., shared libs)
	Offset int64
	// Device owner of this mapping (major:minor) in Mkdev format.
	Dev uint64
	// The inode of the device above
	Inode uint64
	// The file or psuedofile (or empty==anonymous)
	Pathname string
}

type file struct {
	dev   uint64 `river:"dev,attr,optional"`
	inode uint64 `river:"inode,attr,optional"`
	path  string `river:"path,attr,optional"`
}

func (m *ProcMap) file() file {
	return file{
		dev:   m.Dev,
		inode: m.Inode,
		path:  m.Pathname,
	}
}

// parseDevice parses the device token of a line and converts it to a dev_t
// (mkdev) like structure.
func parseDevice(s string) (uint64, error) {
	toks := strings.Split(s, ":")
	if len(toks) < 2 {
		return 0, fmt.Errorf("unexpected number of fields")
	}

	major, err := strconv.ParseUint(toks[0], 16, 0)
	if err != nil {
		return 0, err
	}

	minor, err := strconv.ParseUint(toks[1], 16, 0)
	if err != nil {
		return 0, err
	}

	return unix.Mkdev(uint32(major), uint32(minor)), nil
}

// parseAddress converts a hex-string to a uintptr.
func parseAddress(s string) (uint64, error) {
	a, err := strconv.ParseUint(s, 16, 0)
	if err != nil {
		return 0, err
	}

	return a, nil
}

// parseAddresses parses the start-end address.
func parseAddresses(s string) (uint64, uint64, error) {
	toks := strings.Split(s, "-")
	if len(toks) < 2 {
		return 0, 0, fmt.Errorf("invalid address")
	}

	saddr, err := parseAddress(toks[0])
	if err != nil {
		return 0, 0, err
	}

	eaddr, err := parseAddress(toks[1])
	if err != nil {
		return 0, 0, err
	}

	return saddr, eaddr, nil
}

// parsePermissions parses a token and returns any that are set.
func parsePermissions(s string) (*ProcMapPermissions, error) {
	if len(s) < 4 {
		return nil, fmt.Errorf("invalid permissions token")
	}

	perms := ProcMapPermissions{}
	for _, ch := range s {
		switch ch {
		case 'r':
			perms.Read = true
		case 'w':
			perms.Write = true
		case 'x':
			perms.Execute = true
		case 'p':
			perms.Private = true
		case 's':
			perms.Shared = true
		}
	}

	return &perms, nil
}

// parseProcMap will attempt to parse a single line within a proc/[pid]/maps
// buffer.
func parseProcMap(text string) (*ProcMap, error) {
	fields := strings.Fields(text)
	if len(fields) < 5 {
		return nil, fmt.Errorf("truncated procmap entry: %s", text)
	}

	saddr, eaddr, err := parseAddresses(fields[0])
	if err != nil {
		return nil, err
	}

	perms, err := parsePermissions(fields[1])
	if err != nil {
		return nil, err
	}

	offset, err := strconv.ParseInt(fields[2], 16, 0)
	if err != nil {
		return nil, err
	}

	device, err := parseDevice(fields[3])
	if err != nil {
		return nil, err
	}

	inode, err := strconv.ParseUint(fields[4], 10, 0)
	if err != nil {
		return nil, err
	}

	pathname := ""

	if len(fields) >= 5 {
		pathname = strings.Join(fields[5:], " ")
	}

	return &ProcMap{
		StartAddr: saddr,
		EndAddr:   eaddr,
		Perms:     perms,
		Offset:    offset,
		Dev:       device,
		Inode:     inode,
		Pathname:  pathname,
	}, nil
}

func parseProcMapsExecutableModules(procMaps string) ([]*ProcMap, error) {
	var modules []*ProcMap
	for _, line := range strings.Split(procMaps, "\n") {
		if line == "" {
			continue
		}
		m, err := parseProcMap(line)
		if err != nil {
			return nil, err
		}
		if !m.Perms.Execute {
			continue
		}
		modules = append(modules, m)
	}
	return modules, nil
}
