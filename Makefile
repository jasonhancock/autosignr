all: deps
	go install ./...
deps:
	go get -v ./...
test:
	go test
package:
	$(MAKE) -C packaging
