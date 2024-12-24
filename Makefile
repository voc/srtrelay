.PHONY: srtrelay
srtrelay:
	go build -o srtrelay

install: srtrelay
	mkdir -p $$(pwd)/debian/srtrelay/usr/bin
	install -m 0755 srtrelay $$(pwd)/debian/srtrelay/usr/bin 

