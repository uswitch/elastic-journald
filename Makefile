PROGRAM = elastic-journald
BUILD_NUMBER ?= SNAPSHOT

all: $(PROGRAM)

$(PROGRAM): $(wildcard *.go)
	GO15VENDOREXPERIMENT=1 go build

clean: $(PROGRAM)
	rm -rf $(PROGRAM)

push:
	aws s3 copy $(PROGRAM) s3://uswitch-tools/$(PROGRAM)/$(BUILD_NUMBER)/$(PROGRAM)
