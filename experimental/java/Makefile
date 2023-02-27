include ../../common.mk

OS := $(shell uname | tr '[:upper:]' '[:lower:]')
ARCH := $(shell uname -m)

java:
	gomobile bind -target java/arm64,java/amd64 -v -o repo/nasp/nasp/0.0.0/nasp-0.0.0-$(OS).jar

install-gomobile:
	cd $$HOME && git clone -b plainjava https://github.com/bonifaido/mobile.git
	cd $$HOME/mobile && go install ./cmd/gomobile && go install ./cmd/gobind

install-linux-dependencies:
	sudo apt update
	sudo apt install golang default-jdk git
	sudo apt install gcc-aarch64-linux-gnu gcc-x86-64-linux-gnu

setup-vm: install-linux-dependencies install-gomobile
	grep -qxF 'export PATH=$$PATH:$$HOME/go/bin' ~/.bashrc || echo 'export PATH=$$PATH:$$HOME/go/bin' >> ~/.bashrc
