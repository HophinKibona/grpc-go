all: vet test testrace

build: deps
	go build google.golang.org/grpc-forked/...

clean:
	go clean -i google.golang.org/grpc-forked/...

deps:
	go get -d -v google.golang.org/grpc-forked/...

proto:
	@ if ! which protoc > /dev/null; then \
		echo "error: protoc not installed" >&2; \
		exit 1; \
	fi
	go generate google.golang.org/grpc-forked/...

test: testdeps
	go test -cpu 1,4 -timeout 7m google.golang.org/grpc-forked/...

testsubmodule: testdeps
	cd security/advancedtls && go test -cpu 1,4 -timeout 7m google.golang.org/grpc-forked/security/advancedtls/...

testappengine: testappenginedeps
	goapp test -cpu 1,4 -timeout 7m google.golang.org/grpc-forked/...

testappenginedeps:
	goapp get -d -v -t -tags 'appengine appenginevm' google.golang.org/grpc-forked/...

testdeps:
	go get -d -v -t google.golang.org/grpc-forked/...

testrace: testdeps
	go test -race -cpu 1,4 -timeout 7m google.golang.org/grpc-forked/...

updatedeps:
	go get -d -v -u -f google.golang.org/grpc-forked/...

updatetestdeps:
	go get -d -v -t -u -f google.golang.org/grpc-forked/...

vet: vetdeps
	./vet.sh

vetdeps:
	./vet.sh -install

.PHONY: \
	all \
	build \
	clean \
	deps \
	proto \
	test \
	testappengine \
	testappenginedeps \
	testdeps \
	testrace \
	updatedeps \
	updatetestdeps \
	vet \
	vetdeps
