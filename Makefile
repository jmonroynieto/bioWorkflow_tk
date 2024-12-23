GIT_TAG := $(shell git rev-parse --short HEAD)
.PHONY: build
build_dir := build/
build:
	for tool in saTherapist ; do \
	go build --ldflags="-X main.CommitId=$(GIT_TAG)" -o ${build_dir} ./$$tool ; \
	done
