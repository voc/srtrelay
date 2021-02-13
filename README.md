# srtrelay
Streaming-Relay for the SRT-protocol

Use at your own risk

## Dependencies
**Ubuntu**
  - apt install libsrt1

**Debian 11**:
  - apt install libsrt1-openssl

**Gentoo**:
  - emerge net-libs/srt

## Build
```
go build
```

## Usage
```bash
# start relay
./srtrelay

# start publisher
ffmpeg -i test.mp4 -c copy -f mpegts srt://localhost:1337?streamid=publish/test

# start subscriber
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

### API
See [docs/API.md](docs/API.md) for more information about the API.

## Contributing
See [docts/Contributing.md](docs/Contributing.md)

## Credits
Thanks go to
  - Haivision for [srt](https://github.com/Haivision/srt) and [srtgo](https://github.com/Haivision/srtgo)
  - Edward Wu for [srt-live-server](https://github.com/Edward-Wu/srt-live-server)
  - Quentin Renard for [go-astits](https://github.com/asticode/go-astits)
