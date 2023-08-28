.PHONY: build
build:
	#tinygo build -o kafka_msg_trace_filter.wasm -gc=custom -tags="custommalloc noasm" -scheduler=asyncify -target=wasi ./main.go
	tinygo build -o kafka_msg_trace_filter.wasm -tags="noasm" -scheduler=asyncify -target=wasi ./main.go