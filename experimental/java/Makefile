include ../../common.mk

OS := $(shell uname | tr '[:upper:]' '[:lower:]')
ARCH := $(shell uname -m)

java:
	gomobile bind -target java/arm64,java/amd64 -v -o repo/nasp/nasp/0.0.0/nasp-0.0.0-$(OS).jar

setup-vm:
	sudo apt update
	sudo apt golang default-jdk git
	ifeq ($(ARCH), x86_64)
		sudo apt install gcc-aarch64-linux-gnu
	endif
	ifeq ($(ARCH), aarch64)
		sudo apt install gcc-x86-64-linux-gnu
	endif
	grep -qxF 'export PATH=$$PATH:$$HOME/go/bin' ~/.bashrc || echo 'export PATH=$$PATH:$$HOME/go/bin' >> ~/.bashrc
