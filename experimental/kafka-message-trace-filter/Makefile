include ../../common.mk

.PHONY: build
build:
	tinygo build -o kafka_msg_trace_filter.wasm -tags="noasm" -scheduler=asyncify -target=wasi -no-debug -opt=2 ./main.go