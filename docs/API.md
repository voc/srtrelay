# srtrelay API
See [config.toml.example](../config.toml.example) for configuring the API endpoint.

## Stream status - /streams
Returns a list of active streams with additional statistics.

Content-Type: application/json

Example:
```
GET http://localhost:8080/streams

[{"name":"abc","clients":0,"created":"2020-11-24T23:55:27.265206348+01:00"}]
```