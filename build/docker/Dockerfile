FROM docker.io/library/golang:1.19.7-bullseye@sha256:7767d772324bf9e98417806bfc321a805ab0f41bedcbd1a854929a5e33e29444 as base
WORKDIR /src
COPY ["./src/", "./src/"]

FROM base as builder

COPY /build/swi-k8s-opentelemetry-collector.yaml /src/swi-k8s-opentelemetry-collector.yaml
RUN GO111MODULE=on go install go.opentelemetry.io/collector/cmd/builder@v0.73.0
RUN CGO_ENABLED=0 /go/bin/builder --config ./swi-k8s-opentelemetry-collector.yaml --output-path ./

# create vendor folder (for mend scanning)
SHELL ["/bin/bash", "-c"]
ARG CREATE_VENDOR_DIR
RUN if [[ -z "$CREATE_VENDOR_DIR" ]] ; then echo vendor creation skipped ; else go mod vendor ; fi

FROM base as tests
RUN cd src/processor/swmetricstransformprocessor && go test ./...

FROM alpine:3.17@sha256:e2e16842c9b54d985bf1ef9242a313f36b856181f188de21313820e177002501 as prep
RUN apk --update add ca-certificates

FROM debian:11.6@sha256:6f1f698e20c8ae73b110eb614a2c6d13fbdb504bbbbcf2965c4d0b80b7bb7fb1 as journal
RUN apt update
RUN apt install -y systemd

FROM scratch

ARG USER_UID=10001
USER ${USER_UID}

COPY --from=prep /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /src/swi-k8s-opentelemetry-collector /swi-otelcol

# dynamically linked libraries that are required for journalctl and the journalctl binary itself
#   use `ldd /bin/journalctl` to get dynamically linked libraries from the binary
COPY --from=journal /lib/systemd/libsystemd-shared-247.so /lib/systemd/libsystemd-shared-247.so
COPY --from=journal /lib/x86_64-linux-gnu/libdl.so.2 /lib/x86_64-linux-gnu/libdl.so.2
COPY --from=journal /lib/x86_64-linux-gnu/libc.so.6 /lib/x86_64-linux-gnu/libc.so.6
COPY --from=journal /usr/lib/x86_64-linux-gnu/libacl.so.1 /usr/lib/x86_64-linux-gnu/libacl.so.1
COPY --from=journal /usr/lib/x86_64-linux-gnu/libblkid.so.1 /usr/lib/x86_64-linux-gnu/libblkid.so.1
COPY --from=journal /lib/x86_64-linux-gnu/libcap.so.2 /lib/x86_64-linux-gnu/libcap.so.2
COPY --from=journal /lib/x86_64-linux-gnu/libcrypt.so.1 /lib/x86_64-linux-gnu/libcrypt.so.1
COPY --from=journal /usr/lib/x86_64-linux-gnu/libgcrypt.so.20 /usr/lib/x86_64-linux-gnu/libgcrypt.so.20
COPY --from=journal /usr/lib/x86_64-linux-gnu/libip4tc.so.2 /usr/lib/x86_64-linux-gnu/libip4tc.so.2
COPY --from=journal /usr/lib/x86_64-linux-gnu/libkmod.so.2 /usr/lib/x86_64-linux-gnu/libkmod.so.2
COPY --from=journal /usr/lib/x86_64-linux-gnu/liblz4.so.1 /usr/lib/x86_64-linux-gnu/liblz4.so.1
COPY --from=journal /usr/lib/x86_64-linux-gnu/libmount.so.1 /usr/lib/x86_64-linux-gnu/libmount.so.1
COPY --from=journal /lib/x86_64-linux-gnu/libpam.so.0 /lib/x86_64-linux-gnu/libpam.so.0
COPY --from=journal /lib/x86_64-linux-gnu/librt.so.1 /lib/x86_64-linux-gnu/librt.so.1
COPY --from=journal /usr/lib/x86_64-linux-gnu/libseccomp.so.2 /usr/lib/x86_64-linux-gnu/libseccomp.so.2
COPY --from=journal /lib/x86_64-linux-gnu/libselinux.so.1 /lib/x86_64-linux-gnu/libselinux.so.1
COPY --from=journal /usr/lib/x86_64-linux-gnu/libzstd.so.1 /usr/lib/x86_64-linux-gnu/libzstd.so.1
COPY --from=journal /lib/x86_64-linux-gnu/liblzma.so.5 /lib/x86_64-linux-gnu/liblzma.so.5
COPY --from=journal /lib/x86_64-linux-gnu/libpthread.so.0 /lib/x86_64-linux-gnu/libpthread.so.0
COPY --from=journal /lib64/ld-linux-x86-64.so.2 /lib64/ld-linux-x86-64.so.2
COPY --from=journal /lib/x86_64-linux-gnu/libgpg-error.so.0 /lib/x86_64-linux-gnu/libgpg-error.so.0
COPY --from=journal /usr/lib/x86_64-linux-gnu/libcrypto.so.1.1 /usr/lib/x86_64-linux-gnu/libcrypto.so.1.1
COPY --from=journal /lib/x86_64-linux-gnu/libaudit.so.1 /lib/x86_64-linux-gnu/libaudit.so.1
COPY --from=journal /usr/lib/x86_64-linux-gnu/libpcre2-8.so.0 /usr/lib/x86_64-linux-gnu/libpcre2-8.so.0
COPY --from=journal /lib/x86_64-linux-gnu/libcap-ng.so.0 /lib/x86_64-linux-gnu/libcap-ng.so.0
COPY --from=journal /bin/journalctl /bin/journalctl

ENTRYPOINT ["/swi-otelcol"]
CMD ["--config=/opt/default-config.yaml"]

