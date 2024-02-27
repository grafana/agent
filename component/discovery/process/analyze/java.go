package analyze

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/procfs"
	"golang.org/x/sys/unix"
)

const (
	labelJava            = "__meta_process_java__"
	labelJavaVersion     = "__meta_process_java_version__"
	labelJavaVersionDate = "__meta_process_java_version_date__"
	labelJavaClasspath   = "__meta_process_java_classpath__"
	LabelJavaHome        = "__meta_process_java_home__"
	labelJavaVMFlags     = "__meta_process_java_vm_flags__"
	labelJavaVMType      = "__meta_process_java_vm_type__"
	labelJavaOsName      = "__meta_process_java_os_name__"
	labelJavaOsArch      = "__meta_process_java_os_arch__"
)

type jvmInfo struct {
	classpath       string
	javaHome        string
	javaVersion     string
	javaVersionDate string
	vmFlags         string
	vmType          string
	osName          string
	osArch          string
}

func analyzeJava(input Input, a *Results) error {
	m := a.Labels
	proc, err := procfs.NewProc(int(input.PID))
	if err != nil {
		return err
	}

	executable, err := proc.Executable()
	if err != nil {
		return err
	}
	isJava := false
	if strings.HasSuffix(executable, "java") {
		isJava = true
	} else {
		cmdLine, err := proc.CmdLine()
		if err != nil {
			return err
		}
		for _, c := range cmdLine {
			if strings.HasPrefix(c, "java") {
				isJava = true
				break
			}
		}
	}
	if !isJava {
		return nil
	}

	m[labelJava] = labelValueTrue
	jInfo, err := getInfoFromJcmd(int(input.PID))
	if err != nil {
		return nil
	}
	if jInfo.classpath != "" {
		m[labelJavaClasspath] = jInfo.classpath
	}
	if jInfo.javaHome != "" {
		m[LabelJavaHome] = jInfo.javaHome
	}
	if jInfo.javaVersion != "" {
		m[labelJavaVersion] = jInfo.javaVersion
	}
	if jInfo.javaVersionDate != "" {
		m[labelJavaVersionDate] = jInfo.javaVersionDate
	}
	if jInfo.vmFlags != "" {
		m[labelJavaVMFlags] = jInfo.vmFlags
	}
	if jInfo.vmType != "" {
		m[labelJavaVMType] = jInfo.vmType
	}
	if jInfo.osName != "" {
		m[labelJavaOsName] = jInfo.osName
	}
	if jInfo.osArch != "" {
		m[labelJavaOsArch] = jInfo.osArch
	}
	return nil
}

func getInfoFromJcmd(pid int) (*jvmInfo, error) {
	output, err := attachAndRunJcmdCommand(pid, "VM.system_properties")
	if err != nil {
		return nil, err
	}
	props := strings.Split(output, "\n")
	j := &jvmInfo{
		vmType: "jdk",
	}
	for _, p := range props {
		writeValue(p, "java.home", &j.javaHome)
		writeValue(p, "java.class.path", &j.classpath)
		writeValue(p, "os.name", &j.osName)
		writeValue(p, "os.arch", &j.osArch)
		writeValue(p, "java.version", &j.javaVersion)
		writeValue(p, "java.version.date", &j.javaVersionDate)
	}
	output, err = attachAndRunJcmdCommand(pid, "VM.flags")
	if err != nil {
		return j, nil
	}
	parts := strings.Split(output, "\n")
	if len(parts) > 1 {
		j.vmFlags = parts[1]
	}
	return j, nil
}

func writeValue(p, n string, dest *string) {
	if strings.HasPrefix(p, n+"=") {
		*dest = strings.Trim(p[len(n)+1:], "\"")
	}
}

func attachAndRunJcmdCommand(pid int, cmd string) (string, error) {
	agentUid := uint32(os.Geteuid())
	agentGid := uint32(os.Getegid())
	targetUid, targetGid, nsPid, err := getProcessInfo(pid)
	if err != nil {
		return "", err
	}

	err = enterNS(pid, "net")
	if err != nil {
		return "", err
	}
	err = enterNS(pid, "ipc")
	if err != nil {
		return "", err
	}
	err = enterNS(pid, "mnt")
	if err != nil {
		return "", err
	}

	if (agentGid != targetGid && syscall.Setegid(int(targetGid)) != nil) || (agentUid != targetUid && syscall.Seteuid(int(targetUid)) != nil) {
		return "", errors.New("failed to change credentials to match the target process")
	}

	tmpPath, err := getTmpPath(pid)
	if err != nil {
		return "", err
	}

	signal.Ignore(syscall.SIGPIPE)

	if !checkSocket(nsPid, tmpPath) {
		if err = attachToJvm(pid, nsPid, tmpPath); err != nil {
			return "", err
		}
	}

	fd, err := connectSocket(nsPid, tmpPath)
	if err != nil {
		return "", err
	}
	defer unix.Close(fd)

	return sendRequest(fd, "jcmd", cmd)
}

func getProcessInfo(pid int) (uid, gid uint32, nspid int, err error) {
	path := fmt.Sprintf("/proc/%d/status", pid)
	statusFile, err := os.Open(path)
	if err != nil {
		return 0, 0, 0, err
	}
	defer statusFile.Close()

	scanner := bufio.NewScanner(statusFile)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		switch fields[0] {
		case "Uid:":
			uid64, err := strconv.ParseUint(fields[1], 10, 32)
			if err != nil {
				return 0, 0, 0, err
			}
			uid = uint32(uid64)
		case "Gid:":
			gid64, err := strconv.ParseUint(fields[1], 10, 32)
			if err != nil {
				return 0, 0, 0, err
			}
			gid = uint32(gid64)
		case "NStgid:":
			// PID namespaces can be nested; the last one is the innermost one
			for _, s := range fields[1:] {
				nspid, err = strconv.Atoi(s)
				if err != nil {
					return 0, 0, 0, err
				}
			}
		default:
		}
	}
	return uid, gid, nspid, nil
}

func enterNS(pid int, nsType string) error {
	path := fmt.Sprintf("/proc/%d/ns/%s", pid, nsType)
	selfPath := fmt.Sprintf("/proc/self/ns/%s", nsType)

	var oldNSStat, newNSStat syscall.Stat_t
	if err := syscall.Stat(selfPath, &oldNSStat); err == nil {
		if err := syscall.Stat(path, &newNSStat); err == nil {
			if oldNSStat.Ino != newNSStat.Ino {
				newNS, err := syscall.Open(path, syscall.O_RDONLY, 0)
				_ = syscall.Close(newNS)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func getTmpPath(pid int) (path string, err error) {
	path = fmt.Sprintf("/proc/%d/root/tmp", pid)
	var stats syscall.Stat_t
	return path, syscall.Stat(path, &stats)
}

func checkSocket(pid int, tmpPath string) bool {
	path := fmt.Sprintf("%s/.java_pid%d", tmpPath, pid)

	var stats syscall.Stat_t
	return syscall.Stat(path, &stats) == nil && (stats.Mode&unix.S_IFSOCK) != 0
}

func attachToJvm(pid, nspid int, tmpPath string) error {
	path := fmt.Sprintf("%s/.attach_pid%d", tmpPath, nspid)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, os.FileMode(0660))
	if err != nil {
		return err
	}
	defer file.Close()
	defer os.Remove(path)

	err = syscall.Kill(pid, syscall.SIGQUIT)
	if err != nil {
		return err
	}

	var ts = time.Millisecond * 20
	attached := false
	for !attached && ts.Nanoseconds() < int64(500*time.Millisecond) {
		time.Sleep(ts)
		attached = checkSocket(nspid, tmpPath)
		ts += 20 * time.Millisecond
	}
	return err
}

func connectSocket(pid int, tmpPath string) (int, error) {
	fd, err := unix.Socket(unix.AF_UNIX, unix.SOCK_STREAM, 0)
	if err != nil {
		return -1, err
	}
	addr := unix.SockaddrUnix{
		Name: fmt.Sprintf("%s/.java_pid%d", tmpPath, pid),
	}
	return fd, unix.Connect(fd, &addr)
}

func sendRequest(fd int, cmd string, arg string) (string, error) {
	request := make([]byte, 0, 6+len(cmd)+len(arg))
	request = append(request, byte('1'))
	request = append(request, byte(0))

	request = append(request, []byte(cmd)...)
	request = append(request, byte(0))

	request = append(request, []byte(arg)...)
	request = append(request, []byte{0, 0, 0}...)

	_, err := unix.Write(fd, request)
	if err != nil {
		return "", err
	}

	response := make([]byte, 0)

	buf := make([]byte, 8192)
	n, _ := unix.Read(fd, buf)

	for n != 0 {
		response = append(response, buf...)
		n, err = unix.Read(fd, buf)
	}

	return string(response), err
}
