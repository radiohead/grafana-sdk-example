FROM golang:1.21.1 as builder

WORKDIR /build
COPY go.mod go.sum ./
COPY vendor* ./vendor
RUN test -f vendor/modules.txt || go mod download

COPY cmd cmd
COPY pkg pkg

RUN CGO_ENABLED=0 go build -o "target/operator" cmd/operator/*.go

FROM gcr.io/distroless/static-debian11 AS runtime

COPY --from=builder /build/target/operator /usr/bin/operator

ENTRYPOINT ["/usr/bin/operator"]
