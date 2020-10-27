package color

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"syscall"
)

// Support color:
// 	"TERM=xterm"
// 	"TERM=xterm-vt220"
// 	"TERM=xterm-256color"
// 	"TERM=screen-256color"
// 	"TERM=tmux-256color"
// 	"TERM=rxvt-unicode-256color"
// Don't support color:
// 	"TERM=cygwin"
var specialColorTerms = map[string]bool{
	"screen-256color":       true,
	"tmux-256color":         true,
	"rxvt-unicode-256color": true,
}

// IsConsole Determine whether w is one of stderr, stdout, stdin
func IsConsole(w io.Writer) bool {
	o, ok := w.(*os.File)
	if !ok {
		return false
	}

	fd := o.Fd()

	// fix: cannot use 'o == os.Stdout' to compare
	return fd == uintptr(syscall.Stdout) || fd == uintptr(syscall.Stdin) || fd == uintptr(syscall.Stderr)
}

// IsMSys msys(MINGW64) environment, does not necessarily support color
func IsMSys() bool {
	// like "MSYSTEM=MINGW64"
	if len(os.Getenv("MSYSTEM")) > 0 {
		return true
	}

	return false
}

// IsSupportColor check current console is support color.
//
// Supported:
// 	linux, mac, or windows's ConEmu, Cmder, putty, git-bash.exe
// Not support:
// 	windows cmd.exe, powerShell.exe
func IsSupportColor() bool {
	envTerm := os.Getenv("TERM")
	if strings.Contains(envTerm, "xterm") {
		return true
	}

	// it's special color term
	if _, ok := specialColorTerms[envTerm]; ok {
		return true
	}

	// like on ConEmu software, e.g "ConEmuANSI=ON"
	if os.Getenv("ConEmuANSI") == "ON" {
		return true
	}

	// like on ConEmu software, e.g "ANSICON=189x2000 (189x43)"
	if os.Getenv("ANSICON") != "" {
		return true
	}

	return false
}

// IsSupport256Color render
func IsSupport256Color() bool {
	// "TERM=xterm-256color"
	// "TERM=screen-256color"
	// "TERM=tmux-256color"
	// "TERM=rxvt-unicode-256color"
	return strings.Contains(os.Getenv("TERM"), "256color")
}

// IsSupportTrueColor render. IsSupportRGBColor
func IsSupportTrueColor() bool {
	// "COLORTERM=truecolor"
	return strings.Contains(os.Getenv("COLORTERM"), "truecolor")
}

/*************************************************************
 * print methods(will auto parse color tags)
 *************************************************************/

// Print render color tag and print messages
func Print(a ...interface{}) {
	Fprint(output, a...)
}

// Printf format and print messages
func Printf(format string, a ...interface{}) {
	Fprintf(output, format, a...)
}

// Println messages with new line
func Println(a ...interface{}) {
	Fprintln(output, a...)
}

// Fprint print rendered messages to writer
// Notice: will ignore print error
func Fprint(w io.Writer, a ...interface{}) {
	if isLikeInCmd {
		renderColorCodeOnCmd(func() {
			_, _ = fmt.Fprint(w, Render(a...))
		})
	} else {
		_, _ = fmt.Fprint(w, Render(a...))
	}
}

// Fprintf print format and rendered messages to writer.
// Notice: will ignore print error
func Fprintf(w io.Writer, format string, a ...interface{}) {
	str := fmt.Sprintf(format, a...)
	if isLikeInCmd {
		renderColorCodeOnCmd(func() {
			_, _ = fmt.Fprint(w, ReplaceTag(str))
		})
	} else {
		_, _ = fmt.Fprint(w, ReplaceTag(str))
	}
}

// Fprintln print rendered messages line to writer
// Notice: will ignore print error
func Fprintln(w io.Writer, a ...interface{}) {
	str := formatArgsForPrintln(a)
	if isLikeInCmd {
		renderColorCodeOnCmd(func() {
			_, _ = fmt.Fprintln(w, ReplaceTag(str))
		})
	} else {
		_, _ = fmt.Fprintln(w, ReplaceTag(str))
	}
}

// Lprint passes colored messages to a log.Logger for printing.
// Notice: should be goroutine safe
func Lprint(l *log.Logger, a ...interface{}) {
	if isLikeInCmd {
		renderColorCodeOnCmd(func() {
			l.Print(Render(a...))
		})
	} else {
		l.Print(Render(a...))
	}
}

// Render parse color tags, return rendered string.
// Usage:
//	text := Render("<info>hello</> <cyan>world</>!")
//	fmt.Println(text)
func Render(a ...interface{}) string {
	if len(a) == 0 {
		return ""
	}

	return ReplaceTag(fmt.Sprint(a...))
}

// Sprint parse color tags, return rendered string
func Sprint(args ...interface{}) string {
	return Render(args...)
}

// Sprintf format and return rendered string
func Sprintf(format string, a ...interface{}) string {
	return ReplaceTag(fmt.Sprintf(format, a...))
}

// String alias of the ReplaceTag
func String(s string) string {
	return ReplaceTag(s)
}

// Text alias of the ReplaceTag
func Text(s string) string {
	return ReplaceTag(s)
}

/*************************************************************
 * helper methods for print
 *************************************************************/

// its Win system. linux windows darwin
// func isWindows() bool {
// 	return runtime.GOOS == "windows"
// }

func doPrint(code string, colors []Color, str string) {
	if isLikeInCmd {
		winPrint(str, colors...)
	} else {
		_, _ = fmt.Fprint(output, RenderString(code, str))
	}
}

func doPrintln(code string, colors []Color, args []interface{}) {
	str := formatArgsForPrintln(args)
	if isLikeInCmd {
		winPrintln(str, colors...)
	} else {
		_, _ = fmt.Fprintln(output, RenderString(code, str))
	}
}

func doPrintV2(code, str string) {
	if isLikeInCmd {
		renderColorCodeOnCmd(func() {
			_, _ = fmt.Fprint(output, RenderString(code, str))
		})
	} else {
		_, _ = fmt.Fprint(output, RenderString(code, str))
	}
}

func doPrintlnV2(code string, args []interface{}) {
	str := formatArgsForPrintln(args)
	if isLikeInCmd {
		renderColorCodeOnCmd(func() {
			_, _ = fmt.Fprintln(output, RenderString(code, str))
		})
	} else {
		_, _ = fmt.Fprintln(output, RenderString(code, str))
	}
}

func stringToArr(str, sep string) (arr []string) {
	str = strings.TrimSpace(str)
	if str == "" {
		return
	}

	ss := strings.Split(str, sep)
	for _, val := range ss {
		if val = strings.TrimSpace(val); val != "" {
			arr = append(arr, val)
		}
	}

	return
}

// if use Println, will add spaces for each arg
func formatArgsForPrintln(args []interface{}) (message string) {
	if ln := len(args); ln == 0 {
		message = ""
	} else if ln == 1 {
		message = fmt.Sprint(args[0])
	} else {
		message = fmt.Sprintln(args...)
		// clear last "\n"
		message = message[:len(message)-1]
	}
	return
}
