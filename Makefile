.PHONY: srtrelay
srtrelay:
	CGO_ENABLED=0 go build -o srtrelay

install: srtrelay
	mkdir -p $$(pwd)/debian/srtrelay/usr/bin
	install -m 0755 srtrelay $$(pwd)/debian/srtrelay/usr/bin 

.PHONY: test
test:
	go test ./...

.PHONY: lint
lint:
	golangci-lint run --timeout 5m