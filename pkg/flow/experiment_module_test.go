package flow_test

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	_ "github.com/grafana/agent/component/module/file"
	"github.com/grafana/agent/pkg/flow"
	"github.com/grafana/agent/pkg/flow/internal/testcomponents"
	"github.com/stretchr/testify/require"
)

func TestModuleBug(t *testing.T) {
	testCases := []struct {
		name        string
		module      string
		otherModule string
		config      string
	}{
		{
			name: "TestImportModule",
			module: `
				testcomponents.count "inc" {
                    frequency = "10ms"
                    max = 10
                }

				module.file "default" {
					filename = "other_module"
					arguments {
						input = testcomponents.count.inc.count
					}
				}

				export "output" {
					value = module.file.default.exports.output
				}
			`,
			otherModule: `
				module.string "default" {
					content = ` + strconv.Quote(`argument "input" {
						optional = false
					}
					testcomponents.passthrough "pt" {
						input = argument.input.value
						lag = "5ms"
					}

					module.string "default" {
						content = `+strconv.Quote(`argument "input" {
							optional = false
						}
						testcomponents.passthrough "pt" {
							input = argument.input.value
							lag = "5ms"
						}
		
						export "output" {
							value = testcomponents.passthrough.pt.output
						}`)+`
						arguments {
							input = testcomponents.passthrough.pt.output
						}
					}
	
					export "output" {
						value = module.string.default.exports.output
					}`) + `
					arguments {
						input = argument.input.value
					}
				}
				argument "input" {
					optional = false
				}

				export "output" {
					value = module.string.default.exports.output
				}
			`,
			config: `
				module.file "default" {
					filename = "module"
				}
                testcomponents.summation "sum" {
                    input = module.file.default.exports.output
                }
            `,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filename := "module"
			require.NoError(t, os.WriteFile(filename, []byte(tc.module), 0664))
			defer os.Remove(filename)

			otherFilename := "other_module"
			if tc.otherModule != "" {
				require.NoError(t, os.WriteFile(otherFilename, []byte(tc.otherModule), 0664))
				defer os.Remove(otherFilename)
			}

			ctrl := flow.New(testOptions(t))
			f, err := flow.ParseSource(t.Name(), []byte(tc.config))
			require.NoError(t, err)
			require.NotNil(t, f)

			err = ctrl.LoadSource(f, nil)
			require.NoError(t, err)

			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan struct{})
			go func() {
				ctrl.Run(ctx)
				close(done)
			}()
			defer func() {
				cancel()
				<-done
			}()

			// Check for initial condition
			require.Eventually(t, func() bool {
				export := getExport[testcomponents.SummationExports](t, ctrl, "", "testcomponents.summation.sum")
				return export.LastAdded == 10
			}, 3*time.Second, 10*time.Millisecond)
		})
	}
}
