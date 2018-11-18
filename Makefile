clean:
	rm -rf output/*

fmt:
	gofmt -l -w -s .

build: clean fmt
	go build -v -o output/deployExample main.go

stop:
	-killall -9 "deployExample"

run:
	./output/deployExample