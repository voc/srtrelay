# srtrelay
Streaming-Relay for the SRT-protocol

**EXPERIMENTAL AT BEST, use at your own risk**

## Dependencies
**Debian 10**:
  - apt install libsrt1-openssl

**Gentoo**:
  - emerge net-libs/srt

## Build
```
go build
```

## Usage
```bash
# relay
./srtrelay

# publisher
ffmpeg -i test.mp4 -c copy -f mpegts srt://localhost:1337?streamid=publish/test

# subscriber
ffplay srt://localhost:1337?streamid=play/test
```

### Commandline Flags
```bash
# List available flags
./srtrelay -h
```

### Configuration
Please take a look at [config.toml.example](config.toml.example) to learn more about configuring srtrelay.

The configuration file can be placed under *config.toml* in the current working directory, at */etc/srtrelay/config.toml* or at a custom location specified via the *-config* flag.

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
