FROM docker.io/library/golang:1.24.0-bookworm AS base
WORKDIR /src
COPY ["./src/", "./src/"]

FROM base AS builder

RUN cd src/cmd && CGO_ENABLED=0 GOEXPERIMENT=boringcrypto go build -trimpath -o ../../swi-k8s-opentelemetry-collector "-ldflags=-s -w"

FROM builder AS tests
# run all tests for all go modules in `src` folder
CMD cd src && \
    ret=0 && \
    for d in $(find . -name go.mod -exec dirname {} \;); do \
      (cd "$d" && go test ./...) || ret=1; \
    done && \
    exit $ret
    
FROM debian:12.9 AS journal
RUN apt update
RUN apt install -y systemd
COPY /build/docker/copy-journalbinary.sh /script.sh
RUN chmod +x /script.sh
RUN /script.sh

FROM gcr.io/distroless/static-debian12@sha256:3f2b64ef97bd285e36132c684e6b2ae8f2723293d09aae046196cca64251acac

ARG USER_UID=10001
USER ${USER_UID}

COPY --from=builder /src/swi-k8s-opentelemetry-collector /swi-otelcol
COPY --from=journal /journalctl-deps/ /

ENTRYPOINT ["/swi-otelcol"]
CMD ["--config=/opt/default-config.yaml"]

