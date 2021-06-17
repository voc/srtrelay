# srtrelay ![CI](https://github.com/voc/srtrelay/workflows/CI/badge.svg)
Streaming-Relay for the SRT-protocol

Use at your own risk

## Dependencies
Requires libsrt-1.4.2

**Ubuntu**
  - you will need to [build libsrt yourself](https://github.com/Haivision/srt#build-on-linux)

**Debian 10**:
  - use libsrt-openssl-dev from the [voc repository](https://c3voc.de/wiki/projects:vocbian)
  - or [build it yourself](https://github.com/Haivision/srt#build-on-linux)

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
ffplay -fflags nobuffer srt://localhost:1337?streamid=play/test
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
