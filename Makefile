WORKSPACE=$(shell pwd)

all: deps
	go install ./...
deps:
	go get ./...
test: deps
	go test -v ./...
package:
	$(MAKE) -C packaging
container:
	docker build --no-cache -t builder-autosignr packaging/redhat/
containerrpm:
	docker run -e "BUILD_NUMBER=$(BUILD_NUMBER)" -v $(WORKSPACE):/mnt/build builder-autosignr
