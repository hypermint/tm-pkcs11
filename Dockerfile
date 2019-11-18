
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
  libssl-dev \
  autoconf \
  automake \
  libtool \
  pkg-config \
  git \
  && apt-get clean \
  && rm -rf /var/lib/apt/lists/*

# SoftHSM

RUN git clone https://github.com/opendnssec/SoftHSMv2.git

ENV SOFTHSM_VERSION 2.5.0

RUN cd SoftHSMv2 \
    && git checkout ${SOFTHSM_VERSION} -b ${SOFTHSM_VERSION} \
    && sh autogen.sh \
    && ./configure --prefix=/usr/local \
    && make \
    && make install

# AWS CloudHSM

RUN wget https://s3.amazonaws.com/cloudhsmv2-software/CloudHsmClient/Xenial/cloudhsm-client_latest_amd64.deb
RUN dpkg -i cloudhsm-client_latest_amd64.deb
RUN rm -f cloudhsm-client_latest_amd64.deb

RUN wget https://s3.amazonaws.com/cloudhsmv2-software/CloudHsmClient/Xenial/cloudhsm-client-pkcs11_latest_amd64.deb
RUN dpkg -i cloudhsm-client-pkcs11_latest_amd64.deb
RUN rm -f cloudhsm-client-pkcs11_latest_amd64.deb

# Configure SoftHSM

COPY softhsm2.conf /etc/softhsm2.conf
RUN softhsm2-util --init-token --slot 0 --label "default" --so-pin password --pin password

# Install tm-pkcs11

COPY --from=builder /work/tm-pkcs11 /tm-pkcs11

RUN echo "" > /config.toml

# Setup key for SoftHSM

RUN /tm-pkcs11 genkey

# Entry point

COPY docker-entrypoint.sh /entrypoint.sh
RUN chmod a+x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
