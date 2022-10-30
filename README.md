# Zap Middleware for Echo

This is a very customizable middleware that provides logging and panic recovery for [Echo](https://github.com/labstack/echo) using Uber's [zap](https://github.com/uber-go/zap). Please see [LoggerConfig](https://pkg.go.dev/github.com/tomruk/zap4echo#LoggerConfig) and [RecoverWithConfig](https://pkg.go.dev/github.com/tomruk/zap4echo#RecoverWithConfig) for documentation of configuration fields. Have fun!

## Features

- Highly customizable
    - There's a `Skipper` function so you can skip logging of HTTP requests depending on the `echo.Context`
    - Optionally, only `error`'ed requests can be logged and successful requests can be omitted.
        - By `error`'ed, it is meant that handlers that returned non-nil `error` or responses that has a status code of 3XX, 4XX, or 5XX
    - Custom `msg` field
    - `caller` field is not logged by default — because it is redundant — but it can be enabled with `IncludeCaller`
    - You can omit certain log fields for convenience or performance reasons.
    - Request IDs are logged. Custom header name can be set with `CustomRequestIDHeader`
    - Custom log fields can be added depending on the `echo.Context` using `FieldAdder` function.
    - Errors given as function argument to `panic` can be handled with `ErrorHandler`
    - Logging of the stack trace can be customized.
- Convenient and quick to use
- Performant
    - zap4echo is designed to be performant.
    - Echo and zap is one of the most performant framework/logger combination in the Go ecosystem.
- Well tested
    - %97.9 test coverage at the time of writing.

## Fields Logged

Please note that in addition, extra fields can be added with `FieldAdder` function.

- Logger
    - `proto` - Protocol
    - `host` - Host header
    - `method` - HTTP method
    - `status` - Status as integer
    - `response_size` - Size of the HTTP response
    - `latency` - Time passed between the start and end of handling the request
    - `status_text` - HTTP status as text
    - `client_ip` - Client IP address
    - `user_agent` - User agent
    - `path` - URL path
    - `request_id` - Request ID (Uses `echo.HeaderXRequestID` by default. Custom header can be set with `CustomRequestIDHeader`)
    - `referer` - Referer
- Recover
    - `error` - Error of the panic
    - `method` - HTTP method
    - `path` - URL path
    - `client_ip` - Client IP address
    - `stacktrace` (if enabled)
    - `request_id` - Request ID (Uses `echo.HeaderXRequestID` by default. Custom header can be set with `CustomRequestIDHeader`)

## Usage

This is a quick cheat sheet. For complete examples, have a look at: [basic](_examples/basic/main.go) and [full](_examples/full/main.go)

```shell
go get -u github.com/tomruk/zap4echo
```

```go
log, _ := zap.NewDevelopment()

e.Use(
    zap4echo.Logger(log),
    zap4echo.Recover(log),
)
```

Then curl it:
```shell
curl http://host:port
```

### Customization

Configuration for customization are documented with comments. For documentation, have a look at [LoggerConfig](https://pkg.go.dev/github.com/tomruk/zap4echo#LoggerConfig) and [RecoverWithConfig](https://pkg.go.dev/github.com/tomruk/zap4echo#RecoverWithConfig)

```go
log, _ := zap.NewDevelopment()

e.Use(zap4echo.LoggerWithConfig(log, zap4echo.LoggerConfig{
    // Configure here...
}))

e.Use(zap4echo.RecoverWithConfig(log, zap4echo.RecoverConfig{
    // Configure here...
}))
```
