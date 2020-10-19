# srtrelay
Streaming-Relay for the SRT-protocol

**EXPERIMENTAL AT BEST, use at your own risk**

## Usage
```bash
# relay
./srtrelay

# publisher
ffmpeg -i test.mp4 -c copy -f mpegts srt://localhost:8090?streamid=test/publish

# subscriber
ffplay srt://localhost:8090?streamid=test/play
```

## Design Ideas
  - Just a 1:n relay, one publisher (push), multiple subscribers (pull)
  - No decoding -> use ffmpeg instead
  - No remuxing -> use ffmpeg instead
  - Allow any data to be relayed, not just MPEG-TS

## Develop
Run tests
```
go test ./...
```

## Credits
Thanks go to
  - Edward Wu for [srt-live-server](https://github.com/Edward-Wu/srt-live-server)
  - Quentin Renard for [go-astits](https://github.com/asticode/go-astits)
