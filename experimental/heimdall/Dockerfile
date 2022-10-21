FROM golang as build

WORKDIR /go/src/app
COPY . .
WORKDIR /go/src/app/experimental/heimdall

RUN go mod download

RUN CGO_ENABLED=0 go build -o /go/bin/app

FROM gcr.io/distroless/static:latest

COPY --from=build /go/bin/app /
#ENV GIN_MODE=release
CMD ["/app"]
