# SPDX-License-Identifier: MIT
# SPDX-FileCopyrightText: © 2022 Uri Shaked <uri@wokwi.com>

ZIP_VERSION ?= unknown
GIT_SHA =  $(shell git rev-parse HEAD)
BUILD_TIME = $(shell date -u --rfc-2822)
GO_FLAGS = -ldflags "-X 'main.gitHash=$(GIT_SHA)' -X 'main.buildTime=$(BUILD_TIME)' -X 'main.version=$(ZIP_VERSION)'"

.PHONY: default
default: build

run:
	go run ./cmd/wokwigw

.PHONY: build
build:
	GOOS=windows              go build $(GO_FLAGS) -o bin/wokwigw.exe ./cmd/wokwigw
	GOOS=darwin               go build $(GO_FLAGS) -o bin/wokwigw-darwin ./cmd/wokwigw
	GOOS=darwin  GOARCH=arm64 go build $(GO_FLAGS) -o bin/wokwigw-darwin_arm64 ./cmd/wokwigw
	GOOS=linux                go build $(GO_FLAGS) -o bin/wokwigw-linux ./cmd/wokwigw
	GOOS=linux   GOARCH=arm64 go build $(GO_FLAGS) -o bin/wokwigw-linux_arm64 ./cmd/wokwigw

.PHONY: zip
zip: build
	zip -9j bin/wokwigw_$(ZIP_VERSION)_Windows_64bit.zip bin/wokwigw.exe LICENSE
	zip -9j bin/wokwigw_$(ZIP_VERSION)_macOS_64bit.zip bin/wokwigw-darwin LICENSE
	zip -9j bin/wokwigw_$(ZIP_VERSION)_macOS_ARM64.zip bin/wokwigw-darwin_arm64 LICENSE
	zip -9j bin/wokwigw_$(ZIP_VERSION)_Linux_64bit.zip bin/wokwigw-linux LICENSE
	zip -9j bin/wokwigw_$(ZIP_VERSION)_Linux_ARM64.zip bin/wokwigw-linux_arm64 LICENSE

.PHONY: test
test:
	go test -v	./...

.PHONY: clean
clean:
	rm -rf bin/wokwigw*
