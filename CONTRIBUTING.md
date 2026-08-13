# Contributing

Thanks for considering!
I'm happy to receive PRs.

## Setup

Install [pre-commit](https://pre-commit.com/#installation) hooks.

```bash
pre-commit install
```

Install dependencies:

```bash
go mod tidy
```

## Tests

Run tests:

```bash
go test ./... -test.short
```

Running the [examples](https://github.com/dratasich/thingsboard-go-client-examples)
with an own ThingsBoard instance provide the integration tests sometimes needed
(to check with TB behavior).

## Update

Update all dependencies:

```bash
go get -u
```
