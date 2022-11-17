include ../../common.mk

dynamic-lib:
	go build -o nasp/lib/nasp.so -buildmode=c-shared nasp/nasp.go

static-lib:
	go build -o nasp/lib/nasp.a -buildmode=c-archive nasp/nasp.go
