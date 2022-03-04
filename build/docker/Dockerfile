ARG GOLANG_IMAGE=docker.io/library/golang:1.17.1-alpine@sha256:13919fb9091f6667cb375d5fdf016ecd6d3a5d5995603000d422b04583de4ef9

FROM ${GOLANG_IMAGE} as builder

COPY /build/nighthawk-swi-opentelemetry-collector.yaml /src/nighthawk-swi-opentelemetry-collector.yaml

RUN GO111MODULE=on go install go.opentelemetry.io/collector/cmd/builder@v0.45.0

WORKDIR /src
RUN CGO_ENABLED=0 /go/bin/builder --config ./nighthawk-swi-opentelemetry-collector.yaml --output-path ./

FROM alpine:latest as prep
RUN apk --update add ca-certificates

FROM scratch

ARG USER_UID=10001
USER ${USER_UID}

COPY --from=prep /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /src/nighthawk-swi-opentelemetry-collector /nighthawk-swi-opentelemetry-collector

ENTRYPOINT ["/nighthawk-swi-opentelemetry-collector"]

