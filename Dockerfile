FROM golang:1.13.4-alpine3.10 as builder

WORKDIR /work

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build

FROM alpine:3.10

COPY --from=builder /work/tm-pkcs11 /usr/local/bin/tm-pkcs11
