PROGRAM = elastic-journald

BUILD_NUMBER ?= SNAPSHOT-$(shell git rev-parse --short HEAD)

all: $(PROGRAM)

$(PROGRAM): $(wildcard *.go)
	GO15VENDOREXPERIMENT=1 go build

clean: $(PROGRAM)
	rm -rf $(PROGRAM)

push: $(PROGRAM)
	aws s3 cp $(PROGRAM) s3://uswitch-tools/$(PROGRAM)/$(BUILD_NUMBER)/$(PROGRAM)


docker-image: Dockerfile
	docker build -t elastic-journald-build .

docker-build: docker-image
	docker run \
	-u $(shell id -u):$(shell id -g) \
	-w /opt/go/src/github.com/uswitch/elastic-journald \
	-v $(PWD):/opt/go/src/github.com/uswitch/elastic-journald \
	elastic-journald-build \
	build
