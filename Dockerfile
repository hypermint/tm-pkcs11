
FROM golang:1.13.4-alpine3.10 as builder

WORKDIR /work

COPY go.mod go.sum ./
RUN go mod download
COPY . .

# http://tsujitaku50.hatenablog.com/entry/2017/09/26/193342
RUN CGO_ENABLED=0 go build -a -tags netgo -installsuffix netgo --ldflags '-extldflags "-static"'

FROM ubuntu:16.04

RUN apt-get update && apt-get install -y \
  bash \
  wget \
  libedit2 \
  libjson-c2 \
  python \
  && apt-get clean \
  && rm -rf /var/lib/apt/lists/*

RUN wget https://s3.amazonaws.com/cloudhsmv2-software/CloudHsmClient/Xenial/cloudhsm-client_latest_amd64.deb
RUN dpkg -i cloudhsm-client_latest_amd64.deb

RUN wget https://s3.amazonaws.com/cloudhsmv2-software/CloudHsmClient/Xenial/cloudhsm-client-pkcs11_latest_amd64.deb
RUN dpkg -i cloudhsm-client-pkcs11_latest_amd64.deb

COPY --from=builder /work/tm-pkcs11 /tm-pkcs11

COPY entrypoint.sh /entrypoint.sh
RUN chmod a+x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
