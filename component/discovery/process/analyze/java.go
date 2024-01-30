package analyze

import (
	"bufio"
	"errors"
	"io"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/prometheus/procfs"
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

func analyzeJava(pid string, reader io.ReaderAt, m map[string]string) error {
	pidn, _ := strconv.Atoi(pid)
	proc, err := procfs.NewProc(pidn)
	if err != nil {
		return err
	}

	cmdLine, err := proc.CmdLine()
	isJava := false
	for _, c := range cmdLine {
		if strings.HasPrefix(c, "java") || strings.HasSuffix(c, "java") {
			isJava = true
			break
		}
	}

	if !isJava {
		return nil
	}
	m[labelJava] = "true"
	jInfo, err := getInfoFromJcmd(pid)
	if err != nil {
		jInfo, _ = getInfoFromReleaseFile(proc)
	}
	if jInfo == nil {
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

func getInfoFromJcmd(pid string) (*jvmInfo, error) {
	cmd := exec.Command("jcmd", pid, "VM.system_properties")
	rawOutput, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	output := string(rawOutput)
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
	cmd = exec.Command("jcmd", pid, "VM.flags")
	rawOutput, err = cmd.CombinedOutput()
	if err != nil {
		return j, nil
	}
	output = string(rawOutput)
	parts := strings.Split(output, "\n")
	if len(parts) > 1 {
		j.vmFlags = parts[1]
	}
	return j, nil
}

func getInfoFromReleaseFile(proc procfs.Proc) (*jvmInfo, error) {
	envVars, err := proc.Environ()
	if err != nil {
		return nil, err
	}
	javaHome := ""
	for _, e := range envVars {
		writeValue(e, "JAVA_HOME", &javaHome)
		if javaHome != "" {
			break
		}
	}

	if javaHome == "" {
		return nil, errors.New("java.home not found")
	}

	file, err := os.Open(path.Join(javaHome, "release"))
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(file)
	j := &jvmInfo{
		javaHome: javaHome,
	}

	for scanner.Scan() {
		p := scanner.Text()
		writeValue(p, "JAVA_VERSION", &j.javaVersion)
		writeValue(p, "JAVA_VERSION_DATE", &j.javaVersionDate)
		writeValue(p, "OS_ARCH", &j.osArch)
		writeValue(p, "OS_NAME", &j.osName)
		writeValue(p, "IMAGE_TYPE", &j.vmType)
	}
	return j, nil
}

func writeValue(p, n string, dest *string) {
	if strings.HasPrefix(p, n+"=") {
		*dest = strings.Trim(p[len(n)+1:], "\"")
	}
}
