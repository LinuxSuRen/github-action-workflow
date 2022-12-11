build:
	go build -o bin/gaw .
copy: build
	cp bin/gaw /usr/local/bin
