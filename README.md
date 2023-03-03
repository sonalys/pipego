# Pipego

Pipego is a robust and type safe pipelining framework, made to improve go's error handling, while also allowing you to load balance, add fail safety, better parallelism and modularization of your \
code.

## Features

This library has support for:

- **Parallelism**: fetch data in parallel, with a cancellable context like in errorGroup implementation.
- **Retriability**: choose from constant, linear and exponential backoffs for retrying any step.
- **Load balance**: you can easily split slices and channels over go-routines using different algorithms.
- **Plug and play api**: you can implement any middleware you want on top of pipego's API.

## Examples

All examples are under the [examples folder](./examples/)

- [Simple pipeline](./examples/simple/main.go)
- [Aggregation](./examples/aggregation/main.go)

## Roadmap

- Add load balancing for channels and maps
- Add Timeout step
