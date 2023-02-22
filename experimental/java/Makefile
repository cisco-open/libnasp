include ../../common.mk

OS := $(shell uname | tr '[:upper:]' '[:lower:]')

java:
	gomobile bind -target java/arm64,java/amd64 -v -o repo/nasp/nasp/0.0.0/nasp-0.0.0-$(OS).jar
