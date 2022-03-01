.PHONY: default fpng
export GOPATH:=${HOME}/go

default: build/bin/fpng

build/bin/fpng:
	go build -o build/bin/fpng main.go


