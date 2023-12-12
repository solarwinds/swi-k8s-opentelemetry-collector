FROM docker.io/library/golang:1.21.5-bookworm as base
WORKDIR /src
COPY ["./src/", "./src/"]

FROM base as builder

COPY /build/swi-k8s-opentelemetry-collector.yaml /src/swi-k8s-opentelemetry-collector.yaml
RUN GO111MODULE=on go install go.opentelemetry.io/collector/cmd/builder@v0.81.0
RUN CGO_ENABLED=0 GOEXPERIMENT=boringcrypto /go/bin/builder --config ./swi-k8s-opentelemetry-collector.yaml --output-path ./

# create vendor folder (for mend scanning)
SHELL ["/bin/bash", "-c"]
ARG CREATE_VENDOR_DIR
RUN if [[ -z "$CREATE_VENDOR_DIR" ]] ; then echo vendor creation skipped ; else go mod vendor ; fi

FROM base as tests
RUN cd src/processor/swmetricstransformprocessor && go test ./...

FROM debian:12.2 as journal
RUN apt update
RUN apt install -y systemd
COPY /build/docker/copy-journalbinary.sh /script.sh
RUN chmod +x /script.sh
RUN /script.sh

FROM base as wrapper
WORKDIR /src/src/wrapper
RUN CGO_ENABLED=0 GOEXPERIMENT=boringcrypto go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o /bin/wrapper && chmod +x /bin/wrapper

FROM alpine:3.19
RUN apk --update add ca-certificates

ARG USER_UID=10001
USER ${USER_UID}

COPY --from=builder /src/swi-k8s-opentelemetry-collector /swi-otelcol
COPY --from=wrapper /bin/wrapper /wrapper
COPY --from=journal /journalctl-deps/ /

ENTRYPOINT ["/wrapper"]
CMD ["/swi-otelcol", "--config=/opt/default-config.yaml"]

