# Contributing
Please run tests before creating a pull request
```
go test ./...
```

## General concepts
  - srtrelay hsould just be a 1:n multiplexer, one publisher (push) to multiple subscribers (pull)
  - Don't try to reimplement functionality already present elsewhere in the stack (e.g. remuxing/transcoding)
  - Allow any data to be relayed, not just MPEG-TS