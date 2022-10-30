# Zap Middleware for Echo

This is a full featured middleware that provides logging and panic recovery for the [Echo framework](https://github.com/labstack/echo) using Uber's [zap](https://github.com/uber-go/zap).

## Features

- Highly customizable
- Simple but extendable
- Well tested

## Usage

This is a quick cheat sheet. For examples, see: [basic](_examples/basic/main.go) and [full](_examples/full/main.go)

```shell
go get -v github.com/tomruk/zap4echo
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
curl http://127.0.0.1:8000
```

## Customization

Configuration for customization are documented via comments. For documentation, see: [LoggerConfig](https://pkg.go.dev/github.com/tomruk/zap4echo#LoggerConfig) and [RecoverWithConfig](https://pkg.go.dev/github.com/tomruk/zap4echo#RecoverWithConfig).

```go
log, _ := zap.NewDevelopment()

e.Use(zap4echo.LoggerWithConfig(log, zap4echo.LoggerConfig{
    // Configure here...
}))

e.Use(zap4echo.RecoverWithConfig(log, zap4echo.RecoverConfig{
    // Configure here...
}))
```
