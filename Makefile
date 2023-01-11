build:
	go build -o bin/gaw .
test:
	go test ./...
pre-commit: test build

copy: build
	cp bin/gaw /usr/local/bin
test-gh:
	act -W pkg/data/ -j imageTest
image:
	docker build . -t ghcr.io/linuxsuren/github-action-workflow:dev --build-arg GOPROXY=https://goproxy.io,direct
push-image: image
	docker push ghcr.io/linuxsuren/github-action-workflow:dev
