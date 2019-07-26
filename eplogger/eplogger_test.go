package eplogger

import (
	"context"
	"errors"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"testing"
	"time"
)

const stringFieldKey = "StringField"

// AppendKeyvalserTest is a type used for testing
type AppendKeyvalserTest struct {
	StringField string
}

// AppendKeyvals implements AppendKeyvalser
func (l AppendKeyvalserTest) AppendKeyvals(keyvals []interface{}) []interface{} {
	return append(keyvals, stringFieldKey, l.StringField)
}

func (l AppendKeyvalserTest) Keyvals() []interface{} {
	return []interface{}{stringFieldKey, l.StringField}
}

// TestLoggingMiddleware tests the logging middleware to ensure
// the underlying endpoing is called and that data is logged as
// expected.
func TestLoggingMiddleware(t *testing.T) {
	var tests = map[string]struct {
		expectLevel           level.Value
		expectKVCount         int
		req                   interface{}
		resp                  interface{}
		inRespErr             error
		reqIsAppendKeyvalser  bool
		respIsAppendKeyvalser bool
	}{
		"nil error": {
			expectLevel:           level.InfoValue(),
			expectKVCount:         10,
			req:                   AppendKeyvalserTest{StringField: "req string"},
			resp:                  AppendKeyvalserTest{StringField: "resp string"},
			inRespErr:             nil,
			reqIsAppendKeyvalser:  true,
			respIsAppendKeyvalser: true,
		},
		"non-nil error": {
			expectLevel:           level.ErrorValue(),
			expectKVCount:         10,
			req:                   AppendKeyvalserTest{StringField: "req string"},
			resp:                  AppendKeyvalserTest{StringField: "resp string"},
			inRespErr:             errors.New("an error"),
			reqIsAppendKeyvalser:  true,
			respIsAppendKeyvalser: true,
		},
		"nil error, no AppendKeyvalser": {
			expectLevel:           level.InfoValue(),
			expectKVCount:         6,
			req:                   "req string",
			resp:                  "resp string",
			inRespErr:             nil,
			reqIsAppendKeyvalser:  false,
			respIsAppendKeyvalser: false,
		},
		"non-nil error, no AppendKeyvalser": {
			expectLevel:           level.ErrorValue(),
			expectKVCount:         6,
			req:                   "req string",
			resp:                  "resp string",
			inRespErr:             errors.New("an error"),
			reqIsAppendKeyvalser:  false,
			respIsAppendKeyvalser: false,
		},
		"nil error, req only AppendKeyvalser": {
			expectLevel:           level.InfoValue(),
			expectKVCount:         8,
			req:                   AppendKeyvalserTest{StringField: "req string"},
			resp:                  "resp string",
			inRespErr:             nil,
			reqIsAppendKeyvalser:  false,
			respIsAppendKeyvalser: false,
		},
		"non-nil error, req only AppendKeyvalser": {
			expectLevel:           level.ErrorValue(),
			expectKVCount:         8,
			req:                   AppendKeyvalserTest{StringField: "req string"},
			resp:                  "resp string",
			inRespErr:             errors.New("an error"),
			reqIsAppendKeyvalser:  false,
			respIsAppendKeyvalser: false,
		},
		"nil error, resp only AppendKeyvalser": {
			expectLevel:           level.InfoValue(),
			expectKVCount:         8,
			req:                   "req string",
			resp:                  AppendKeyvalserTest{StringField: "resp string"},
			inRespErr:             nil,
			reqIsAppendKeyvalser:  false,
			respIsAppendKeyvalser: false,
		},
		"non-nil error, resp only AppendKeyvalser": {
			expectLevel:           level.ErrorValue(),
			expectKVCount:         8,
			req:                   "req string",
			resp:                  AppendKeyvalserTest{StringField: "resp string"},
			inRespErr:             errors.New("an error"),
			reqIsAppendKeyvalser:  false,
			respIsAppendKeyvalser: false,
		},
	}

	// Run the sub-tests
	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			// setup the logger
			var output []interface{}
			logger := log.Logger(log.LoggerFunc(func(keyvals ...interface{}) error {
				output = keyvals
				return nil
			}))
			errLogger := level.Error(logger)
			logger = level.Info(logger)

			// Simulate a go-kit endpoint
			endpointExecuted := false
			ep := func(ctx context.Context, request interface{}) (response interface{}, err error) {
				endpointExecuted = true
				return tt.resp, tt.inRespErr
			}

			// Wrap the simulated endpoint with the middleware
			epWithMw := LoggingMiddleware(logger, errLogger)(ep)

			// Execute the endpoint and middleware
			epWithMw(context.Background(), tt.req)

			if len(output) != tt.expectKVCount {
				t.Errorf("len of output is different than expected: want %d, have %d", tt.expectKVCount, len(output))
			}
			if !endpointExecuted {
				t.Errorf("endpoint was never executed")
			}
			if want, have := "level", output[0]; want != have {
				t.Errorf("output[0]: want %s, have %s", want, have)
			}
			if want, have := tt.expectLevel, output[1]; want != have {
				t.Errorf("output[1]: want %s, have %s", want, have)
			}
			if want, have := transErrKey, output[2]; want != have {
				t.Errorf("output[2]: want %s, have %s", want, have)
			}
			if want, have := tt.inRespErr, output[3]; want != have {
				t.Errorf("output[3]: want %s, have %s", want, have)
			}
			if want, have := tookKey, output[4]; want != have {
				t.Errorf("output[4]: want %s, have %s", want, have)
			}
			_, ok := output[5].(time.Duration)
			if !ok {
				t.Fatalf("output[5]: want time.Time, have %T", output[5])
			}

			if tt.reqIsAppendKeyvalser && tt.respIsAppendKeyvalser {
				if want, have := stringFieldKey, output[6]; want != have {
					t.Errorf("output[6]: want %s, have %s", want, have)
				}
				req, ok := tt.req.(AppendKeyvalserTest)
				if !ok {
					t.Error("tt.req does not implement AppendKeyvalserTest")
				}
				if want, have := req.StringField, output[7]; want != have {
					t.Errorf("output[7]: want %s, have %s", want, have)
				}

				if want, have := stringFieldKey, output[8]; want != have {
					t.Errorf("output[8]: want %s, have %s", want, have)
				}
				resp, ok := tt.resp.(AppendKeyvalserTest)
				if !ok {
					t.Error("tt.resp does not implement AppendKeyvalserTest")
				}
				if want, have := resp.StringField, output[9]; want != have {
					t.Errorf("output[9]: want %s, have %s", want, have)
				}
			} else if tt.reqIsAppendKeyvalser {
				if want, have := stringFieldKey, output[6]; want != have {
					t.Errorf("output[6]: want %s, have %s", want, have)
				}
				req, ok := tt.req.(AppendKeyvalserTest)
				if !ok {
					t.Error("tt.req does not implement AppendKeyvalserTest")
				}
				if want, have := req.StringField, output[7]; want != have {
					t.Errorf("output[7]: want %s, have %s", want, have)
				}
			} else if tt.respIsAppendKeyvalser {
				if want, have := stringFieldKey, output[6]; want != have {
					t.Errorf("output[6]: want %s, have %s", want, have)
				}
				resp, ok := tt.resp.(AppendKeyvalserTest)
				if !ok {
					t.Error("tt.resp does not implement AppendKeyvalserTest")
				}
				if want, have := resp.StringField, output[7]; want != have {
					t.Errorf("output[7]: want %s, have %s", want, have)
				}
			}
		})
	}
}

// BenchmarkLoggingMiddlewareWithErr tests how long the middleware takes to execute when
// the resulting err is not nil.
// The benchmark output by BenchmarkLoggingMiddlewareCreation should be subtracted from the
// benchmark output by this func.
func BenchmarkLoggingMiddleware(b *testing.B) {
	benchmarks := map[string]struct {
		err  error
		req  interface{}
		resp interface{}
	}{
		"req:AppendKeyvalserTest,resp:AppendKeyvalserTest,error:nil": {
			err:  nil,
			req:  AppendKeyvalserTest{StringField: "test req"},
			resp: AppendKeyvalserTest{StringField: "test resp"},
		},
		"req:AppendKeyvalserTest,resp:AppendKeyvalserTest,error:non-nil": {
			err:  errors.New("an error"),
			req:  AppendKeyvalserTest{StringField: "test req"},
			resp: AppendKeyvalserTest{StringField: "test resp"},
		},
		"req:string,resp:string,error:nil": {
			err:  nil,
			req:  "test req",
			resp: "test resp",
		},
		"req:string,resp:string,error:non-nil": {
			err:  errors.New("an error"),
			req:  "test req",
			resp: "test resp",
		},
		"req:AppendKeyvalserTest,resp:string,error:nil": {
			err:  nil,
			req:  AppendKeyvalserTest{StringField: "test req"},
			resp: "test resp",
		},
		"req:AppendKeyvalserTest,resp:string,error:non-nil": {
			err:  errors.New("an error"),
			req:  AppendKeyvalserTest{StringField: "test req"},
			resp: "test resp",
		},
		"req:string,resp:AppendKeyvalserTest,error:nil": {
			err:  nil,
			req:  "test req",
			resp: AppendKeyvalserTest{StringField: "test resp"},
		},
		"req:string,resp:AppendKeyvalserTest,error:non-nil": {
			err:  errors.New("an error"),
			req:  "test req",
			resp: AppendKeyvalserTest{StringField: "test resp"},
		},
	}

	for testName, bb := range benchmarks {
		b.Run(testName, func(b *testing.B) {
			ctx := context.Background()
			req := bb.req
			resp := bb.resp

			// Simulate a go-kit endpoint
			ep := func(ctx context.Context, request interface{}) (interface{}, error) {
				return resp, bb.err
			}

			// Wrap the simulated endpoint with the middleware. Need to do this for each
			// b.N iteration so that a new logger instance can be passed to the middleware
			// to avoid memory copying from slowing the benchmark down during log.With()
			logger := log.NewNopLogger()
			logger, errLogger := level.Info(logger), level.Error(logger)
			epWithMw := LoggingMiddleware(logger, errLogger)(ep)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Execute the endpoint and middleware
				epWithMw(ctx, req)
			}
		})
	}
}

// BenchmarkLoggingMiddlewareCreation outputs how long initialization takes
// of the LoggingMiddleware. The benchmark returned here can be subtracted from
// other benchmarks to get an accurate representation of their purposes.
func BenchmarkLoggingMiddlewareCreation(b *testing.B) {
	resp := AppendKeyvalserTest{
		StringField: "resp string",
	}

	// Simulate a go-kit endpoint
	epErr := errors.New("an error")
	ep := func(ctx context.Context, request interface{}) (response interface{}, err error) {
		return resp, epErr
	}

	logger := log.NewNopLogger()
	logger, errLogger := level.Info(logger), level.Error(logger)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Wrap the simulated endpoint with the middleware. Need to do this for each
		// b.N iteration so that a new logger instance can be passed to the middleware
		// to avoid memory copying from slowing the benchmark down during log.With()
		LoggingMiddleware(logger, errLogger)(ep)
	}
}

// BenchmarkMakeKeyvals tests how long it takes to add all logging
// key/values into a single keyvals []interface
func BenchmarkMakeKeyvals(b *testing.B) {
	req := AppendKeyvalserTest{
		StringField: "req string",
	}
	resp := AppendKeyvalserTest{
		StringField: "resp string",
	}
	d := time.Duration(1)
	err := errors.New("test")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		makeKeyvals(req, resp, d, err)
	}
}
