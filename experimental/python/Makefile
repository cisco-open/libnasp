include ../../common.mk

dynamic-lib:
	go build -o nasp.so -buildmode=c-shared nasp.go

static-lib:
	go build -o nasp.a -buildmode=c-archive nasp.go
