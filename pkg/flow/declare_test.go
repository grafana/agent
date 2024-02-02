package flow_test

import (
	"context"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/grafana/agent/pkg/flow"
	"github.com/grafana/agent/pkg/flow/internal/testcomponents"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/grafana/agent/service"
	"github.com/stretchr/testify/require"
)

type testCase struct {
	name     string
	config   string
	expected int
}

func TestDeclare(t *testing.T) {
	tt := []testCase{
		{
			name: "BasicDeclare",
			config: `
			declare "test" {
				argument "input" {
					optional = false
				}
			
				testcomponents.passthrough "pt" {
					input = argument.input.value
					lag = "1ms"
				}
			
				export "output" {
					value = testcomponents.passthrough.pt.output
				}
			}
			testcomponents.count "inc" {
				frequency = "10ms"
				max = 10
			}
		
			test "myModule" {
				input = testcomponents.count.inc.count
			}
		
			testcomponents.summation "sum" {
				input = test.myModule.output
			}
			`,
			expected: 10,
		},
		{
			name: "NestedDeclares",
			config: `
			declare "test" {
				argument "input" {
					optional = false
				}
				declare "nested" {
					argument "input" {
						optional = false
					}
					export "output" {
						value = argument.input.value
					}
				}
			
				testcomponents.passthrough "pt" {
					input = argument.input.value
					lag = "1ms"
				}
				nested "default" {
					input = testcomponents.passthrough.pt.output
				}
			
				export "output" {
					value = nested.default.output
				}
			}
			testcomponents.count "inc" {
				frequency = "10ms"
				max = 10
			}
		
			test "myModule" {
				input = testcomponents.count.inc.count
			}
		
			testcomponents.summation "sum" {
				input = test.myModule.output
			}
			`,
			expected: 10,
		},
		{
			name: "DeclaredInParentDepth1",
			config: `
			declare "test" {
				argument "input" {
					optional = false
				}
			
				testcomponents.passthrough "pt" {
					input = argument.input.value
					lag = "1ms"
				}
				rootDeclare "default" {
					input = testcomponents.passthrough.pt.output
				}
			
				export "output" {
					value = rootDeclare.default.output
				}
			}
			declare "rootDeclare" {
				argument "input" {
					optional = false
				}
				export "output" {
					value = argument.input.value
				}
			}
			testcomponents.count "inc" {
				frequency = "10ms"
				max = 10
			}
		
			test "myModule" {
				input = testcomponents.count.inc.count
			}
		
			testcomponents.summation "sum" {
				input = test.myModule.output
			}
			`,
			expected: 10,
		},
		{
			name: "DeclaredInParentDepth2",
			config: `
			declare "test" {
				argument "input" {
					optional = false
				}
			
				testcomponents.passthrough "pt" {
					input = argument.input.value
					lag = "1ms"
				}
				declare "anotherDeclare" {
					argument "input" {
						optional = false
					}
					rootDeclare "default" {
						input = argument.input.value
					}
					export "output" {
						value = rootDeclare.default.output
					}
				}
				anotherDeclare "myOtherDeclare" {
					input = testcomponents.passthrough.pt.output
				}
			
				export "output" {
					value = anotherDeclare.myOtherDeclare.output
				}
			}
			declare "rootDeclare" {
				argument "input" {
					optional = false
				}
				export "output" {
					value = argument.input.value
				}
			}
			testcomponents.count "inc" {
				frequency = "10ms"
				max = 10
			}
		
			test "myModule" {
				input = testcomponents.count.inc.count
			}
		
			testcomponents.summation "sum" {
				input = test.myModule.output
			}
			`,
			expected: 10,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
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

			require.Eventually(t, func() bool {
				export := getExport[testcomponents.SummationExports](t, ctrl, "", "testcomponents.summation.sum")
				return export.LastAdded == tc.expected
			}, 3*time.Second, 10*time.Millisecond)
		})
	}
}

type errorTestCase struct {
	name          string
	config        string
	expectedError *regexp.Regexp
}

func TestDeclareError(t *testing.T) {
	tt := []errorTestCase{
		{
			name: "CircleDependencyBetweenDeclares",
			config: `
			declare "a" {
				b "t1" {}
			}
			declare "b" {
				a "t2" {}
			}
			a "t3" {}
			`,
			// using regex here because the order of the node can vary
			// not ideal because it could technically match "a" and "a"
			expectedError: regexp.MustCompile(`cycle: declare\.(a|b), declare\.(a|b)`),
		},
		{
			name: "CircleDependencyWithinDeclare",
			config: `
			declare "a" {
				declare "b" {
					c "t1" {}
				}
				declare "c" {
					b "t2" {}
				}
				b "t3" {}
			}
			a "t4" {}
			`,
			expectedError: regexp.MustCompile(`cycle: declare\.(b|c), declare\.(b|c)`),
		},
		{
			name: "CircleDependencyWithItself",
			config: `
			declare "a" {
				a "t1" {}
			}
			a "t2" {}
			`,
			expectedError: regexp.MustCompile(`self reference: declare\.a`),
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			defer verifyNoGoroutineLeaks(t)
			s, err := logging.New(os.Stderr, logging.DefaultOptions)
			require.NoError(t, err)
			ctrl := flow.New(flow.Options{
				Logger:   s,
				DataPath: t.TempDir(),
				Reg:      nil,
				Services: []service.Service{},
			})
			f, err := flow.ParseSource(t.Name(), []byte(tc.config))
			require.NoError(t, err)
			require.NotNil(t, f)

			err = ctrl.LoadSource(f, nil)
			if !tc.expectedError.MatchString(err.Error()) {
				t.Errorf("Expected error to match regex %q, but got: %v", tc.expectedError, err)
			}

			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan struct{})
			go func() {
				ctrl.Run(ctx)
				close(done)
			}()
			cancel()
			<-done
		})
	}
}
