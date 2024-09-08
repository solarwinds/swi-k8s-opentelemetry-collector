FROM docker.io/library/golang:1.22.6-bookworm AS base
WORKDIR /src
COPY ["./src/", "./src/"]

FROM base AS builder

RUN cd src/cmd && CGO_ENABLED=0 GOEXPERIMENT=boringcrypto go build -trimpath -o ../../swi-k8s-opentelemetry-collector "-ldflags=-s -w"

# create vendor folder (for mend scanning)
SHELL ["/bin/bash", "-c"]
ARG CREATE_VENDOR_DIR
RUN if [[ -z "$CREATE_VENDOR_DIR" ]] ; then echo vendor creation skipped ; else cd src/cmd && go mod vendor -o /src/vendor ; fi

FROM builder AS tests
# run all tests for all go modules in `src` folder
CMD cd src && find . -name go.mod -execdir go test ./... \;

FROM debian:12.6 AS journal
RUN apt update
RUN apt install -y systemd
COPY /build/docker/copy-journalbinary.sh /script.sh
RUN chmod +x /script.sh
RUN /script.sh

FROM base AS wrapper
WORKDIR /src/src/wrapper
RUN CGO_ENABLED=0 GOEXPERIMENT=boringcrypto go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o /bin/wrapper && chmod +x /bin/wrapper

FROM gcr.io/distroless/static-debian12@sha256:95eb83a44a62c1c27e5f0b38d26085c486d71ece83dd64540b7209536bb13f6d

ARG USER_UID=10001
USER ${USER_UID}

COPY --from=builder /src/swi-k8s-opentelemetry-collector /swi-otelcol
COPY --from=wrapper /bin/wrapper /wrapper
COPY --from=journal /journalctl-deps/ /

ENTRYPOINT ["/wrapper"]
CMD ["/swi-otelcol", "--config=/opt/default-config.yaml"]

