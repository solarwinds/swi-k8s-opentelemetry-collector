FROM docker.io/library/golang:1.23.3-bookworm AS base
WORKDIR /src
COPY ["./src/", "./src/"]

FROM base AS builder

RUN cd src/cmd && CGO_ENABLED=0 GOEXPERIMENT=boringcrypto go build -trimpath -o ../../swi-k8s-opentelemetry-collector "-ldflags=-s -w"

FROM builder AS tests
# run all tests for all go modules in `src` folder
CMD cd src && find . -name go.mod -execdir go test ./... \;

FROM debian:12.8 AS journal
RUN apt update
RUN apt install -y systemd
COPY /build/docker/copy-journalbinary.sh /script.sh
RUN chmod +x /script.sh
RUN /script.sh

FROM base AS wrapper
WORKDIR /src/src/wrapper
RUN CGO_ENABLED=0 GOEXPERIMENT=boringcrypto go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o /bin/wrapper && chmod +x /bin/wrapper

FROM gcr.io/distroless/static-debian12@sha256:f4a57e8ffd7ba407bdd0eb315bb54ef1f21a2100a7f032e9102e4da34fe7c196

ARG USER_UID=10001
USER ${USER_UID}

COPY --from=builder /src/swi-k8s-opentelemetry-collector /swi-otelcol
COPY --from=wrapper /bin/wrapper /wrapper
COPY --from=journal /journalctl-deps/ /

ENTRYPOINT ["/wrapper"]
CMD ["/swi-otelcol", "--config=/opt/default-config.yaml"]

