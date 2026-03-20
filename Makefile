IMAGE ?= quay.io/matzew/kaudy:latest

.PHONY: build install container push test clean

build:
	go build -o kaudy ./cmd/kaudy/

install:
	go install ./cmd/kaudy/

container:
	podman build -t $(IMAGE) .

push: container
	podman push $(IMAGE)

test:
	go test ./...

clean:
	rm -f kaudy
