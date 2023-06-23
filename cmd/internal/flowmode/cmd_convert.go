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
)

func convertCommand() *cobra.Command {
	f := &flowConvert{
		write:          false,
		sourceFormat:   "",
		bypassWarnings: false,
	}

	cmd := &cobra.Command{
		Use:   "convert [flags] file",
		Short: "Convert a supported config file to River",
		Long: `The convert subcommand translates a supported config file to
a River configuration file.

If the file argument is not supplied or if the file argument is "-", then convert will read from stdin.

The -w flag can be used to write the formatted file back to disk. Output will be written to SOURCE_FILEPATH.river. -w can not be provided when convert is reading from stdin. When -w is not provided, fmt will write the result to stdout.

The -f flag can be used to specify the format we are converting from.

The -b flag can be used to bypass warnings.`,
		Args:         cobra.RangeArgs(0, 3),
		SilenceUsage: true,
		Aliases:      []string{"convert"},

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

	cmd.Flags().BoolVarP(&f.write, "write", "w", f.write, "Write the converted file back to disk when not reading from standard input.")
	cmd.Flags().StringVarP(&f.sourceFormat, "source-format", "f", f.sourceFormat, "The format of the source file. Supported formats: 'prometheus'.")
	cmd.Flags().BoolVarP(&f.bypassWarnings, "bypass-warnings", "b", f.bypassWarnings, "Enable bypassing warnings when converting")
	return cmd
}

type flowConvert struct {
	write          bool
	sourceFormat   string
	bypassWarnings bool
}

func (fc *flowConvert) Run(configFile string) error {
	switch configFile {
	case "-":
		if fc.write {
			return fmt.Errorf("cannot use -w with standard input")
		}
		return convert("<stdin>", nil, os.Stdin, fc)

	default:
		fi, err := os.Stat(configFile)
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return fmt.Errorf("cannot convert a directory")
		}

		f, err := os.Open(configFile)
		if err != nil {
			return err
		}
		defer f.Close()
		return convert(configFile, fi, f, fc)
	}
}

func convert(filename string, fi os.FileInfo, r io.Reader, fc *flowConvert) error {
	bb, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	var diags convert_diag.Diagnostics
	bb, diags = converter.Convert(bb, converter.Input(fc.sourceFormat))
	hasErrors := diags.HasErrorLevel(convert_diag.SeverityLevelError)
	hasWarns := diags.HasErrorLevel(convert_diag.SeverityLevelWarn)
	if hasErrors || (!fc.bypassWarnings && hasWarns) {
		return diags
	}
	buf.WriteString(string(bb))

	if !fc.write {
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
