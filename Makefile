build:
	go build -o bin/gaw .
copy: build
	cp bin/gaw /usr/local/bin
test-gh:
	act -W pkg/data/ -j imageTest
image:
	docker build .
