.PHONY: build run

GOFILES = $(shell find . -name '*.go')

build: catnip-fyne

run: catnip-fyne
	./catnip-fyne

catnip-fyne: $(GOFILES)
	go build -v -tags wayland
