# Contributing
Run tests before creating a pull request
```
go test ./...
```

## General concepts
  - Just a 1:n multiplexer, one publisher (push) to multiple subscribers (pull)
  - Don't try to reimplement functionality already present elsewhere in the stack (e.g. remuxing/transcoding)
  - Allow any data to be relayed, not just MPEG-TS

## TODOs
### Figure out what's going on with ffmpeg srt support
If an srt stream is read with ffmpeg (atleast for h.264 + MPEG-TS)