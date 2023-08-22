module github.com/cisco/nasp/kafka-message-trace-filter

go 1.19

replace (
	github.com/tetratelabs/proxy-wasm-go-sdk => ../../../proxy-wasm-go-sdk
	wwwin-github.cisco.com/eti/kafka-protocol-go/pkg => ../../../kafka-protocol-go/pkg
	wwwin-github.cisco.com/eti/kafka-protocol-go => ../../../kafka-protocol-go
)

require (
	github.com/stretchr/testify v1.8.2
	github.com/tetratelabs/proxy-wasm-go-sdk v0.0.0-00010101000000-000000000000
	wwwin-github.cisco.com/eti/kafka-protocol-go v0.0.0-20230705090120-80b8ffe3e9a4
)

require (
	emperror.dev/errors v0.8.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20230111030713-bf00bc1b83b6 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/klauspost/compress v1.16.5 // indirect
	github.com/pierrec/lz4/v4 v4.1.17 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/tetratelabs/wazero v1.0.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	wwwin-github.cisco.com/eti/kafka-protocol-go/pkg v0.0.0 // indirect
)
