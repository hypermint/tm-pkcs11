
FROM golang:1.13.4-stretch as builder

WORKDIR /work

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build

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
