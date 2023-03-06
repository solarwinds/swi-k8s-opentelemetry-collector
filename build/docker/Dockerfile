ARG GOLANG_IMAGE=docker.io/library/golang:1.19.3-bullseye@sha256:d388153691a825844ebb3586dd04d1c60a2215522cc445701424205dffc8a83e

FROM ${GOLANG_IMAGE} as base
WORKDIR /src
COPY ["./src/", "./src/"]

FROM base as builder

COPY /build/swi-k8s-opentelemetry-collector.yaml /src/swi-k8s-opentelemetry-collector.yaml
RUN GO111MODULE=on go install go.opentelemetry.io/collector/cmd/builder@v0.69.0
RUN CGO_ENABLED=0 /go/bin/builder --config ./swi-k8s-opentelemetry-collector.yaml --output-path ./

# create vendor folder (for mend scanning)
SHELL ["/bin/bash", "-c"]
ARG CREATE_VENDOR_DIR
RUN if [[ -z "$CREATE_VENDOR_DIR" ]] ; then echo vendor creation skipped ; else go mod vendor ; fi

FROM base as tests
RUN cd src/processor/swmetricstransformprocessor && go test ./...

FROM alpine:latest as prep
RUN apk --update add ca-certificates

FROM debian:11.4 as journal
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
