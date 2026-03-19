IMAGE ?= quay.io/matzew/kaudy:latest

.PHONY: build install container test clean

build:
	go build -o kaudy ./cmd/kaudy/

install:
	go install ./cmd/kaudy/

container:
	podman build -t $(IMAGE) .

test:
	go test ./...

clean:
	rm -f kaudy
