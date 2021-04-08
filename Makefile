srtrelay:
	HOME=$$(pwd) git config --global http.sslVerify false
	mkdir -p $$(pwd)/gopath
	HOME=$$(pwd) GOPATH=$$(pwd)/gopath go build -o srtrelay

install: talkiepi
	mkdir -p $$(pwd)/debian/srtrelay/usr/local/bin
	install -m 0755 srtrelay $$(pwd)/debian/srtrelay/usr/local/bin 

