# syntax=docker/dockerfile:1
FROM golang:latest

RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

RUN apt-get update && \
    apt-get install -y git tclsh pkg-config cmake libssl-dev build-essential ninja-build && \
    git clone --depth 1 --branch v1.4.3 https://github.com/Haivision/srt.git libsrt && \
    cmake -S libsrt -B libsrt-build -G Ninja && \
    ninja -C libsrt-build && \
    ninja -C libsrt-build install && \
    rm -rf libsrt libsrt-build && \
    apt-get purge -y git tclsh pkg-config cmake libssl-dev build-essential ninja-build && \
    apt-get autoremove -y && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

ADD . /build
WORKDIR /build
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -v -o srtrelay .

# clean start
FROM ubuntu:latest
RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y libssl1.1 && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*
RUN useradd -l -m -r srtrelay
USER srtrelay
WORKDIR /home/srtrelay/
COPY --from=0 /build/config.toml.example ./config.toml
COPY --from=0 /build/srtrelay ./
COPY --from=0 /usr/local/lib/libsrt* /usr/lib/
CMD ["./srtrelay"]
