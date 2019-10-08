TAG := "1.0.0"

build:
	CGO_ENABLED=0 go build -o docker/bindata cmd/bindata/*.go

push:
	cd docker && docker build -t f1shl3gs/bindata:${TAG} .
	docker push f1shl3gs/bindata:${TAG}
