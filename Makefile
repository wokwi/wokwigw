# SPDX-License-Identifier: MIT
# SPDX-FileCopyrightText: © 2022 Uri Shaked <uri@wokwi.com>

.PHONY: default
default: build

run:
	go run ./cmd/wokwigw

.PHONY: build
build:
	GOOS=windows              go build -o bin/wokwigw.exe ./cmd/wokwigw
	GOOS=darwin               go build -o bin/wokwigw-darwin ./cmd/wokwigw
	GOOS=darwin  GOARCH=arm64 go build -o bin/wokwigw-darwin_arm64 ./cmd/wokwigw
	GOOS=linux                go build -o bin/wokwigw-linux ./cmd/wokwigw
	GOOS=linux   GOARCH=arm64 go build -o bin/wokwigw-linux_arm64 ./cmd/wokwigw

.PHONY: test
test:
	go test -v	./...

.PHONY: clean
clean:
	rm -rf bin/wokwigw*
