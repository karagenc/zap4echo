# Zap Middleware for Echo

<div align="left">
    <a href="https://github.com/karagenc/zap4echo/actions/workflows/tests.yml"><img src="https://github.com/karagenc/zap4echo/actions/workflows/tests.yml/badge.svg"></img></a>
    <a href="https://coveralls.io/github/karagenc/zap4echo"><img src="https://coveralls.io/repos/github/karagenc/zap4echo/badge.svg"></img></a>
    <a href="https://pkg.go.dev/github.com/karagenc/zap4echo"><img src="https://pkg.go.dev/badge/github.com/karagenc/zap4echo"></img></a>
</div><br>

This is a very customizable middleware that provides logging and panic recovery for [Echo](https://github.com/labstack/echo) using Uber's [zap](https://github.com/uber-go/zap). Please see [LoggerConfig](https://pkg.go.dev/github.com/karagenc/zap4echo#LoggerConfig) and [RecoverConfig](https://pkg.go.dev/github.com/karagenc/zap4echo#RecoverConfig) for documentation of configuration fields. Have fun!

## Features

- Highly customizable
    - There's a `Skipper` function so you can skip logging of HTTP requests depending on the `echo.Context`
    - Error only logging: Logging can be limited to requests resulted in error — requests which either returned an error or has a status code of 3xx, 4xx, or 5xx.
    - Custom `msg` field
    - `caller` field is not logged by default. Logging can be enabled with `IncludeCaller`.
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
    - 100% test coverage

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
go get -u github.com/karagenc/zap4echo
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

Configuration for customization are documented with comments. For documentation, have a look at [LoggerConfig](https://pkg.go.dev/github.com/karagenc/zap4echo#LoggerConfig) and [RecoverConfig](https://pkg.go.dev/github.com/karagenc/zap4echo#RecoverConfig)

```go
log, _ := zap.NewDevelopment()

e.Use(zap4echo.LoggerWithConfig(log, zap4echo.LoggerConfig{
    // Configure here...
}))

e.Use(zap4echo.RecoverWithConfig(log, zap4echo.RecoverConfig{
    // Configure here...
}))
```
