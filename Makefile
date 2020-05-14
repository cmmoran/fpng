.PHONY: default fpng
export GOPATH:=${HOME}/go

default: fpng

fpng:
	go install -a fpng.go


