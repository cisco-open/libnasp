## Kafka protocol deserializer/serializer library in Go

This is a Go library for deserializing/serializing raw Kafka messages. The library supports TinyGo and WASI platform through TinyGo.  

## Go versions

Requires go version 1.19 and above

## Kafka versions

Tested with Kafka versions 2.8.1 to 3.1.0

### Installation

```shell
go get github.com/cisco-open/nasp/components/kafka-protocol-go/pkg
```

### Consuming the library

```go
package main

import (
    "github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/request"
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/response"
)

func main() {
	var rawMsg []byte // Kafka raw message; should not include the 4 bytes length message size
	
	// kafka request message
	req, err := request.Parse(rawMsg)
	if err != nil {
		...
    }
	defer req.Release()
	...
	
	// kafka response message
	resp, err := response.Parse(rawMsg, requestApiKey, requestApiVersion, requestCorrelationId)
	if err != nil {
		...
    }
	defer resp.Release()
	
}


```

## Development

### Kafka message deserializers/serializers generation

In order to avoid the use of reflection the Kafka message serializers/deserializers code which reside under `kafka-protocol-go/pkg/protocol/messages` folder are generated using the Kafka message descriptors located at `kafka-protocol-go/assets` folder.

Run `make generate` to regenerate Kafka message serializers/deserializers code.

### Kafka message Go structs field alignment optimisation

To optimize the memory footprint of the Kafka message Go structs run `make fieldalignment`. This should always be invoked in case the generated Kafka message Go structs are changed.

### Kafka protocol version update

The code that generates the Kafka message deserializers/serializers resides under the `kafka-protocol-go/cmd` and uses the Kafka protocol descriptors located under `kafka-protocol-go/assets/protocol_spec`.
Download the new Kafka protocol descriptors from https://github.com/apache/kafka/tree/trunk/clients/src/main/resources/common/message to the `kafka-protocol-go/assets/protocol_spec` folder and then regenerate the deserializers/serializers.

In case new types/concepts are introduced in the Kafka protocol the generator under `kafka-protocol-go/cmd` must be updated accordingly.


### Escape analysis

Run `make escape-analysis` to generate heap escape analysis report

### Tests

Run `make test` to execute unit tests and generate code coverage report

### Benchmarks

Run `make benchmark` to generate stats on execution times and heap allocations