include ../../common.mk

OS := $(shell uname | tr '[:upper:]' '[:lower:]')
ARCH := $(shell uname -m)

java:
	gomobile bind -target java/arm64,java/amd64 -v -o repo/nasp/nasp/0.0.0/nasp-0.0.0-$(OS).jar

install-gomobile:
	mkdir -p $(shell go env GOPATH)/src/gomobile-bonifaido
	git clone -b plainjava https://github.com/bonifaido/mobile.git $(shell go env GOPATH)/src/gomobile-bonifaido
	cd $(shell go env GOPATH)/src/gomobile-bonifaido && go install ./cmd/gomobile && go install ./cmd/gobind

install-linux-dependencies:
	sudo apt update
	sudo snap install go --classic
	sudo apt install default-jdk git
	sudo apt install gcc-aarch64-linux-gnu gcc-x86-64-linux-gnu

setup-vm: install-linux-dependencies install-gomobile
	grep -qxF 'export PATH=$$PATH:$$HOME/go/bin' ~/.bashrc || echo 'export PATH=$$PATH:$$HOME/go/bin' >> ~/.bashrc
