VERSION := "1.1.0"

PHONY: build

build:
	CGO_ENABLED=0 go build -ldflags "-X main.version=${VERSION}" -o bindata cmd/bindata/*.go

push: build
	docker build -f docker/Dockerfile -t f1shl3gs/bindata:${VERSION} .
	docker push f1shl3gs/bindata:${VERSION}
