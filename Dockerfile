
FROM golang:1.13.4-alpine3.10 as builder

WORKDIR /work

COPY go.mod go.sum ./
RUN go mod download
COPY . .

# http://tsujitaku50.hatenablog.com/entry/2017/09/26/193342
RUN CGO_ENABLED=0 go build -a -tags netgo -installsuffix netgo --ldflags '-extldflags "-static"'

FROM library/amazonlinux:2.0.20191016.0

COPY --from=builder /work/tm-pkcs11 /usr/local/bin/tm-pkcs11

RUN yum install -y wget

RUN wget https://s3.amazonaws.com/cloudhsmv2-software/CloudHsmClient/EL7/cloudhsm-client-latest.el7.x86_64.rpm
RUN yum install -y ./cloudhsm-client-latest.el7.x86_64.rpm
