# eplogger

An endpoint middleware that provides logging of every request with request and response type-specific fields 
logged based upon the implementation of the `AppendKeyvalser` interface. 

### Problem

Due to the wonderful flexibility of go-kit the endpoint layer uses the `interface{}` type for both the request and response
object types. This means that there is no ability at the endpoint layer to know the specific implementation type, and therefore, no
ability to access type-specific field values.

### Solution

Define an interface that can provide additional context at the endpoint layer for logging using [kit/log](https://github.com/go-kit/kit/tree/master/log):

```go
type AppendKeyvalser interface {
	AppendKeyvals(keyvals []interface{}) []interface{}
}
```

Implement the `AppendKeyvalser` interface on each request or the response type that should have type-specific values logged.

### Getting Started

Implement the `AppendKeyvalser` interface on a request/response type that should have fields logged within the eplogger
middleware:

```go
// SumRequest collects the request parameters for the Sum method.
type SumRequest struct {
	A int `json:"a"`
	B int `json:"b"`
}

// AppendKeyvals implements eplogger.AppendKeyvalser to return keyvals specific to SumRequest for logging
func (s SumRequest) AppendKeyvals(keyvals []interface{}) []interface{} {
	return append(keyvals,
		"SumRequest.A", s.A,
		"SumRequest.B", s.B)
}
```

Now the eplogger package ([eplogger.go](eplogger.go)) can be used as a go-kit middleware to take advantage of the AppendKeyvalser
interface:

```go
// Create a logger that will be used for normal logging and another for error logging
errLogger := level.Error(logger)
logger = level.Info(logger)

// Wrap an endpoint with the eplogger middleware
// errLogger is optional. If errLogger is nil then logger will be used for all events.
epWithMw := eplogger.LoggingMiddleware(logger, errLogger)(endpoint)
```

Now, every time a request passes through an endpoint, keyvals specific to `request` and `response` will be logged
if the `AppendKeyvalser` interface is implemented. If the interface is not implemented then the log will just be
written without any information about the `request` and `response` objects. For example (eplogger added the keys that start with `SumRequest.`, `SumResponse.`, `MultiplyRequest.`, and `MultiplyResponse.`):

```bash
ts=2019-06-12T20:01:41.671115Z caller=level.go:150 method=Sum level=info SumRequest.A=2 SumRequest.B=3 SumResponse.R=5 SumResponse.Err=null transport_error=null took=11.08µs
ts=2019-06-12T20:01:45.862347Z caller=level.go:150 method=Multiply level=info MultiplyRequest.A=2 MultiplyRequest.B=3 MultiplyResponse.R=6 MultiplyResponse.Err=null transport_error=null took=6.548µs
```

#### Example

You can view a working example [here](https://github.com/jwenz723/gokit-example).