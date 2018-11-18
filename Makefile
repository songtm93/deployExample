clean:
	rm -rf output/*

fmt:
	gofmt -l -w -s .

install: fmt
    go install -v main.go
    go build  -v -o output/deployExample main.go

stop:
    killall -9 "deployExample"

run: install
    ./output/deployExample