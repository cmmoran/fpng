.PHONY: default server client deps fmt clean all transmit release-all assets client-assets server-assets contributors
export GOPATH:=/Users/cm250309/.go

BUILDTAGS=debug
default: all

fpng:
	go install -a -tags '$(BUILDTAGS)' fpng

