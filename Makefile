WORKSPACE=$(shell pwd)
goversion=1.13.7

all:
	go install ./...

test:
	go test -v ./...

package:
	$(MAKE) -C packaging

container:
	docker build --no-cache -t builder-autosignr --build-arg goversion=$(goversion) packaging/redhat/

containerrpm:
	docker run \
		-e "BUILD_NUMBER=$(BUILD_NUMBER)" \
		-e "GOPROXY=$(GOPROXY)" \
		-v $(WORKSPACE):/mnt/build \
		builder-autosignr
