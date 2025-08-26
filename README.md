# thingsboard-go-client-sdk

![build](https://github.com/dratasich/thingsboard-go-client-sdk/actions/workflows/go.yml/badge.svg)

ThingsBoard go client SDK supporting:

* Unencrypted and encrypted (TLS v1.2) connection via MQTT
* Device MQTT API:
  * Listen and handle RPCs

## Contributing

Install [pre-commit](https://pre-commit.com/#installation) hooks.
```bash
pre-commit install
```

Install dependencies:
```bash
go mod tidy
```

Run tests:
```bash
go test ./... -test.short
```

I'm happy to receive PRs.
