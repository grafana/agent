package flowmode

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"

	"github.com/grafana/agent/converter"
	convert_diag "github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/pkg/river/diag"
	"github.com/grafana/agent/pkg/river/parser"
	"github.com/grafana/agent/pkg/river/printer"
)

func fmtCommand() *cobra.Command {
	f := &flowFmt{
		write:                 false,
		convertSourceFormat:   "",
		convertBypassWarnings: false,
	}

	cmd := &cobra.Command{
		Use:   "fmt [flags] file",
		Short: "Format a River file",
		Long: `The fmt subcommand applies standard formatting rules to the specified
configuration file. It can format an existing river file or convert support config
formats to river.

If the file argument is not supplied or if the file argument is "-", then fmt will read from stdin.

The -w flag can be used to write the formatted file back to disk. Output will be written to FILEPATH.river. -w can not be provided when fmt is reading from stdin. When -w is not provided, fmt will write the result to stdout.

The -f flag can be used to specify that we are converting from a format other than river.

The -b flag can be used to specify we should bypass warnings when converting from a format other than river.`,
		Args:         cobra.RangeArgs(0, 3),
		SilenceUsage: true,
		Aliases:      []string{"format"},

		RunE: func(_ *cobra.Command, args []string) error {
			var err error

			if len(args) == 0 {
				// Read from stdin when there are no args provided.
				err = f.Run("-")
			} else {
				err = f.Run(args[0])
			}

			var diags diag.Diagnostics
			if errors.As(err, &diags) {
				for _, diag := range diags {
					fmt.Fprintln(os.Stderr, diag)
				}
				return fmt.Errorf("encountered errors during formatting")
			}

			return err
		},
	}

	cmd.Flags().BoolVarP(&f.write, "write", "w", f.write, "Write result to a file instead of stdout")
	cmd.Flags().StringVarP(&f.convertSourceFormat, "convert.source-format", "f", f.convertSourceFormat, "The source of the file for reformatting to flow. Only use when translating from a format other than river.  Supported formats: 'prometheus'.")
	cmd.Flags().BoolVarP(&f.convertBypassWarnings, "convert.bypass-warnings", "b", f.convertBypassWarnings, "Enable bypassing warnings during convert")
	return cmd
}

type flowFmt struct {
	write                 bool
	convertSourceFormat   string
	convertBypassWarnings bool
}

func (ff *flowFmt) Run(configFile string) error {
	switch configFile {
	case "-":
		if ff.write {
			return fmt.Errorf("cannot use -w with standard input")
		}
		return format("<stdin>", nil, os.Stdin, ff)

	default:
		fi, err := os.Stat(configFile)
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return fmt.Errorf("cannot format a directory")
		}

		f, err := os.Open(configFile)
		if err != nil {
			return err
		}
		defer f.Close()
		return format(configFile, fi, f, ff)
	}
}

func format(filename string, fi os.FileInfo, r io.Reader, ff *flowFmt) error {
	bb, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if ff.convertSourceFormat != "" {
		var diags convert_diag.Diagnostics
		bb, diags = converter.Convert(bb, converter.Input(ff.convertSourceFormat))
		if diags.HasErrorLevel(convert_diag.SeverityLevelError) ||
			(!ff.convertBypassWarnings && diags.HasErrorLevel(convert_diag.SeverityLevelWarn)) {
			return diags
		}
		buf.WriteString(string(bb))
	} else {
		f, err := parser.ParseFile(filename, bb)
		if err != nil {
			return err
		}

		if err := printer.Fprint(&buf, f); err != nil {
			return err
		}

		// Add a newline at the end of the file.
		_, _ = buf.Write([]byte{'\n'})
	}

	if !ff.write {
		_, err := io.Copy(os.Stdout, &buf)
		return err
	}

	filepath := strings.TrimSuffix(filename, path.Ext(filename)) + ".river"
	wf, err := os.OpenFile(filepath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, fi.Mode().Perm())
	if err != nil {
		return err
	}
	defer wf.Close()

	_, err = io.Copy(wf, &buf)
	return err
}
