FROM docker.io/library/golang:1.19.0-alpine as builder
RUN apk add make git
WORKDIR /workspace
COPY . .
RUN make build && mv bin/notation /usr/bin/notation

FROM docker.io/library/alpine:3.15.4
COPY --from=builder /usr/bin/notation /usr/bin/notation
CMD [ "/usr/bin/notation" ]
