package eplogger

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"time"
)

// AppendKeyvalser is an interface that wraps the basic AppendKeyvals method.
//
// AppendKeyvals should be implemented to append key/value pairs into keyvals
// without removing any existing elements, then return the extended keyvals.
//		Example:
//			// Define your struct type
//			type SomeType struct{
//				AField string
//				BField string
//			}
//
//			// Implement the AppendKeyvals func to satisfy the AppendKeyvalser interface
//			func (s SomeType) AppendKeyvals(keyvals []interface{}) []interface{} {
//				// Add key/value sets here (2 values per set, key followed by value)
//			 	return append(keyvals,
//			 		"SomeType.AField", s.AField,
//			 		"SomeType.BField", s.BField)
//			}
type AppendKeyvalser interface {
	AppendKeyvals(keyvals []interface{}) []interface{}
}

const (
	tookKey     = "took"
	transErrKey = "transport_error"
)

// LoggingMiddleware returns an endpoint middleware that logs the
// duration of each invocation, the resulting error (if any), and
// keyvals specific to the request and response object if they implement
// the AppendKeyvalser interface.
//
// The level specified as defaultLevel will be used when the resulting error
// is nil otherwise level.Error will be used.
func LoggingMiddleware(logger, errLogger log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			defer func(begin time.Time) {
				kvs := makeKeyvals(request, response, time.Since(begin), err)
				if err != nil {
					errLogger.Log(kvs...)
				} else {
					logger.Log(kvs...)
				}
			}(time.Now())
			return next(ctx, request)
		}
	}
}

// makeKeyvals will place the received parameters into an []interface{} to be
// returned in the order:
// 	1. err
//	2. d
//	3. req (if AppendKeyvalser is implemented)
//	4. resp (if AppendKeyvalser is implemented)
func makeKeyvals(req, resp interface{}, d time.Duration, err error) []interface{} {
	KVs := []interface{}{transErrKey, err, tookKey, d}
	if l, ok := req.(AppendKeyvalser); ok {
		KVs = l.AppendKeyvals(KVs)
	}
	if l, ok := resp.(AppendKeyvalser); ok {
		KVs = l.AppendKeyvals(KVs)
	}
	return KVs
}
