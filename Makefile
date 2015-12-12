all: deps test
	go install ./...
deps:
	go get -v ./...
package:
	$(MAKE) -C packaging
test:
	#cd src/amproxy/amproxy && go test
	#cd src/amproxy/message && go test
	/bin/true
