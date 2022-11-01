include ../../../../common.mk

.PHONY: build-filters
build-filters:
	@find . -type f -name "main.go" \
	| xargs -I {} bash -c 'dirname {}' \
	| xargs -I {} bash -c 'echo "building {}"; tinygo build -o ../{}.wasm -scheduler=none -target=wasi ./{}/main.go'
