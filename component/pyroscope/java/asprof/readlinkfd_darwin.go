package asprof

func readLinkFD(f *os.File) (string, error) {
	fd := f.Fd()

	buf := make([]byte, 4096)
	res, err := unix.FcntlInt(fd, unix.F_GETPATH, int(uintptr(unsafe.Pointer(&buf[0]))))
	if err != nil || res != 0 {
		return "", fmt.Errorf("failed to check fd %d ", fd)
	}
	for i, b := range buf {
		if b == 0 {
			buf = buf[:i]
			break
		}
	}

	return string(buf), nil
}
