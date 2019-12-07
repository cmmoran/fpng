.PHONY: default fpng
export GOPATH:=/Users/cmoran/.go

default: fpng

fpng:
	go install -a fpng.go


