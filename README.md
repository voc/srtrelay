# srtrelay ![CI](https://github.com/voc/srtrelay/workflows/CI/badge.svg)
Streaming-Relay for the SRT-protocol

Use at your own risk.

## Quick start
Run with docker (**Note:** nightly image not recommended for production)
```bash
docker run --rm ghcr.io/voc/srtrelay/srtrelay:latest

# start publisher
ffmpeg -i test.mp4 -c copy -f mpegts srt://localhost:1337?streamid=publish/test

# start subscriber
ffplay -fflags nobuffer srt://localhost:1337?streamid=play/test
```

Start docker with custom config. See [config.toml.example](config.toml.example)
```bash
# provide your own config from the local directory
docker run --rm -v $(pwd)/config.toml:/home/srtrelay/config.toml ghcr.io/voc/srtrelay/srtrelay:latest
```

## Run with docker-compose

In your `docker-compose.yml`:

```yaml
   srtrelay:
     image: ghcr.io/voc/srtrelay/srtrelay:latest
     restart: always
     container_name: srtrelay
     volumes:
       - ./srtrelay-config.toml:/home/srtrelay/config.toml
     ports:
       - "44560:1337/udp"
```

This will forward port `44560` to internal port `1337` in the container. Importantly, forwarding UDP is required.
It will also copy a `srtrelay-config.toml` file in the same directory into the container to use as config.toml

Start the server with the usual

```bash
docker-compose up -d
```

## Build with docker
You will need atleast docker-20.10

```bash
docker build -t srtrelay .

# run srtrelay
docker run --rm -it srtrelay
```

## Build without docker
### Install Dependencies
Requires >=libsrt-1.4.2, golang and a C compiler

**Ubuntu**
  - you will need to [build libsrt yourself](https://github.com/Haivision/srt#build-on-linux)

**Debian 10**:
  - use libsrt-openssl-dev from the [voc repository](https://c3voc.de/wiki/projects:vocbian)
  - or [build it yourself](https://github.com/Haivision/srt#build-on-linux)

**Gentoo**:
  - emerge net-libs/srt

### Build
```bash
go build -o srtrelay

# run srtrelay
./srtrelay
```

## Usage
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
See [docs/Contributing.md](docs/Contributing.md)

## Credits
Thanks go to
  - Haivision for [srt](https://github.com/Haivision/srt) and [srtgo](https://github.com/Haivision/srtgo)
  - Edward Wu for [srt-live-server](https://github.com/Edward-Wu/srt-live-server)
  - Quentin Renard for [go-astits](https://github.com/asticode/go-astits)
