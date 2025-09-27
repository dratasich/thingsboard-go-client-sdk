## v0.5.0 (2025-09-27)

### Feat

- send device's client attributes via gateway
- request device attributes as gateway
- receive attribute updates of connected devices in gateway
- handle device RPCs in gateway
- add GatewayRequestRPC struct
- add connect and send telemetry via gateway

### Fix

- subscription to RPC topic (broke with refactoring for v0.3.0)
- request id of RPCs are integers not strings

### Refactor

- embed base client into gateway
- subscription and handler

## v0.4.0 (2025-09-01)

### Feat

- publish telemetry

### Refactor

- publish calls (attributes, rpc reply)

## v0.3.0 (2025-08-29)

### Feat

- request device attributes
- send and receive attributes

### Refactor

- topic and qos constants

## v0.2.1 (2025-08-29)

### Refactor

- await connection

## v0.2.0 (2025-08-26)

### Feat

- **api**: align tb api reply rpc

## v0.1.1 (2025-08-26)

### Refactor

- align go api to other golang projects

## v0.1.0 (2025-08-26)

### Feat

- **mqtt**: add mqtt connection and rpc processing

### Fix

- package name
- allow go to find our package

### Refactor

- rename model->events, tbmqtt->mqtt
