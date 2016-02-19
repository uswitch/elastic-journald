PROGRAM = elastic-journald
GO_PATH ?= $(shell pwd)
BUILD_NUMBER ?= SNAPSHOT-$(shell git rev-parse --short HEAD)

all: $(PROGRAM)

$(PROGRAM): $(wildcard *.go)
	GO15VENDOREXPERIMENT=1 go build

clean: $(PROGRAM)
	rm -rf $(PROGRAM)

push: $(PROGRAM)
	aws s3 cp $(PROGRAM) s3://uswitch-tools/$(PROGRAM)/$(BUILD_NUMBER)/$(PROGRAM)
