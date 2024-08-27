all: vet test testrace

build:
	go build google.golang.org/grpcv54/...

clean:
	go clean -i google.golang.org/grpcv54/...

deps:
	GO111MODULE=on go get -d -v google.golang.org/grpcv54/...

proto:
	@ if ! which protoc > /dev/null; then \
		echo "error: protoc not installed" >&2; \
		exit 1; \
	fi
	go generate google.golang.org/grpcv54/...

test:
	go test -cpu 1,4 -timeout 7m google.golang.org/grpcv54/...

testsubmodule:
	cd security/advancedtls && go test -cpu 1,4 -timeout 7m google.golang.org/grpcv54/security/advancedtls/...
	cd security/authorization && go test -cpu 1,4 -timeout 7m google.golang.org/grpcv54/security/authorization/...

testrace:
	go test -race -cpu 1,4 -timeout 7m google.golang.org/grpcv54/...

testdeps:
	GO111MODULE=on go get -d -v -t google.golang.org/grpcv54/...

vet: vetdeps
	./vet.sh

vetdeps:
	./vet.sh -install

.PHONY: \
	all \
	build \
	clean \
	proto \
	test \
	testrace \
	vet \
	vetdeps
