# syntax=docker/dockerfile:1
FROM golang:1.24-trixie AS build

RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

COPY . /build
WORKDIR /build
ARG TARGETARCH
RUN GOOS=linux GOARCH=$TARGETARCH go build -v -o srtrelay .

# clean start
FROM debian:trixie
RUN apt-get update && \
    apt-get upgrade -y && \
    rm -rf /var/lib/apt/lists/*
RUN useradd -l -m -r srtrelay
USER srtrelay
WORKDIR /home/srtrelay/
RUN cat <<'EOF' > ./config.toml
[app]
addresses = ["[::]:1337"]
auth = {type = "static", static = {allow = ["*"]}}
EOF
COPY --from=build /build/srtrelay ./
EXPOSE 1337/udp
EXPOSE 8080/tcp
CMD ["./srtrelay"]
