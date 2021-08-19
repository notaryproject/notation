ARG GO_VER=1.16-alpine
ARG DOCKER_VER=20.10

FROM golang:${GO_VER} as go

RUN apk add \
      git \
      make

FROM go as nv2

RUN git clone https://github.com/notaryproject/nv2.git /nv2 \
 && cd /nv2 \
 && make build

FROM go as oras

RUN git clone https://github.com/deislabs/oras.git /oras \
 && cd /oras \
 && git checkout reference-types \
 && make build \
 && cp ./bin/linux/amd64/oras /usr/local/bin/oras

FROM docker:${DOCKER_VER} as docker

FROM go as release

RUN apk add \
      bash \
      curl \
      git \
      jq \
      make \
      openssl

COPY --from=nv2 /nv2/bin/ /usr/local/bin/
COPY --from=oras /usr/local/bin/oras /usr/local/bin/oras
COPY --from=docker /usr/local/bin/docker /usr/local/bin/docker
COPY rebuild.sh /usr/local/bin/

RUN mkdir -p /usr/libexec/docker/cli-plugins/ \
 && ln -s /usr/local/bin/docker-generate /usr/local/bin/docker-nv2 /usr/libexec/docker/cli-plugins/

WORKDIR /root
