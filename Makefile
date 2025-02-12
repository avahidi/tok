
OUT=bin/tok
SRC=./...

build:
	go build -o $(OUT) ./...

test:
	go test -v ./...

bench:
	go test -v -bench=. -run=^Benchmark

run: build
	$(OUT) ls

fmt:
	go fmt ./...

dist:
	GOOS=linux GOARCH=386 go build -o bin/tok-linux-x86 $(SRC)
	GOOS=linux GOARCH=amd64 go build -o bin/tok-linux-x64 $(SRC)
	GOOS=linux GOARCH=arm64 go build -o bin/tok-linux-arm64 $(SRC)
	GOOS=windows GOARCH=386 go build -o bin/tok-win-x86 $(SRC)
	GOOS=windows GOARCH=amd64 go build -o bin/tok-win-x64 $(SRC)
	GOOS=freebsd GOARCH=386 go build -o bin/tok-bsd-x86 $(SRC)
	GOOS=darwin GOARCH=arm64 go build -o bin/tok-osx-arm64 $(SRC)

clean:
	go clean
	rm -rf bin
