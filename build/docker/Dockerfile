FROM docker.io/library/golang:1.21.4-bookworm@sha256:85aacbed94a248f792beb89198649ddbc730649054b397f8d689e9c4c4cceab7 as base
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

FROM alpine:3.18@sha256:eece025e432126ce23f223450a0326fbebde39cdf496a85d8c016293fc851978 as prep
RUN apk --update add ca-certificates

FROM debian:12.2@sha256:133a1f2aa9e55d1c93d0ae1aaa7b94fb141265d0ee3ea677175cdb96f5f990e5 as journal
RUN apt update
RUN apt install -y systemd

FROM base as wrapper
WORKDIR /src/src/wrapper
RUN CGO_ENABLED=0 GOEXPERIMENT=boringcrypto go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o /bin/wrapper && chmod +x /bin/wrapper

FROM scratch as journalbinaries

# dynamically linked libraries that are required for journalctl and the journalctl binary itself
#   use `ldd /bin/journalctl` to get dynamically linked libraries from the binary
COPY --from=journal /bin/journalctl /bin/journalctl
COPY --from=journal /usr/lib/x86_64-linux-gnu/systemd/libsystemd-shared-252.so /usr/lib/x86_64-linux-gnu/systemd/libsystemd-shared-252.so
COPY --from=journal /lib/x86_64-linux-gnu/libacl.so.1 /lib/x86_64-linux-gnu/libacl.so.1
COPY --from=journal /lib/x86_64-linux-gnu/libaudit.so.1 /lib/x86_64-linux-gnu/libaudit.so.1
COPY --from=journal /lib/x86_64-linux-gnu/libblkid.so.1 /lib/x86_64-linux-gnu/libblkid.so.1
COPY --from=journal /lib/x86_64-linux-gnu/libc.so.6 /lib/x86_64-linux-gnu/libc.so.6
COPY --from=journal /lib/x86_64-linux-gnu/libcap-ng.so.0 /lib/x86_64-linux-gnu/libcap-ng.so.0
COPY --from=journal /lib/x86_64-linux-gnu/libcap.so.2 /lib/x86_64-linux-gnu/libcap.so.2
COPY --from=journal /lib/x86_64-linux-gnu/libcrypt.so.1 /lib/x86_64-linux-gnu/libcrypt.so.1
COPY --from=journal /lib/x86_64-linux-gnu/libcrypto.so.3 /lib/x86_64-linux-gnu/libcrypto.so.3
COPY --from=journal /lib/x86_64-linux-gnu/libgcrypt.so.20 /lib/x86_64-linux-gnu/libgcrypt.so.20
COPY --from=journal /lib/x86_64-linux-gnu/libgpg-error.so.0 /lib/x86_64-linux-gnu/libgpg-error.so.0
COPY --from=journal /lib/x86_64-linux-gnu/libip4tc.so.2 /lib/x86_64-linux-gnu/libip4tc.so.2
COPY --from=journal /lib/x86_64-linux-gnu/libkmod.so.2 /lib/x86_64-linux-gnu/libkmod.so.2
COPY --from=journal /lib/x86_64-linux-gnu/liblz4.so.1 /lib/x86_64-linux-gnu/liblz4.so.1
COPY --from=journal /lib/x86_64-linux-gnu/liblzma.so.5 /lib/x86_64-linux-gnu/liblzma.so.5
COPY --from=journal /lib/x86_64-linux-gnu/libm.so.6 /lib/x86_64-linux-gnu/libm.so.6
COPY --from=journal /lib/x86_64-linux-gnu/libmount.so.1 /lib/x86_64-linux-gnu/libmount.so.1
COPY --from=journal /lib/x86_64-linux-gnu/libpam.so.0 /lib/x86_64-linux-gnu/libpam.so.0
COPY --from=journal /lib/x86_64-linux-gnu/libpcre2-8.so.0 /lib/x86_64-linux-gnu/libpcre2-8.so.0
COPY --from=journal /lib/x86_64-linux-gnu/libseccomp.so.2 /lib/x86_64-linux-gnu/libseccomp.so.2
COPY --from=journal /lib/x86_64-linux-gnu/libselinux.so.1 /lib/x86_64-linux-gnu/libselinux.so.1
COPY --from=journal /lib/x86_64-linux-gnu/libzstd.so.1 /lib/x86_64-linux-gnu/libzstd.so.1
COPY --from=journal /lib64/ld-linux-x86-64.so.2 /lib64/ld-linux-x86-64.so.2

FROM scratch

ARG USER_UID=10001
USER ${USER_UID}

COPY --from=prep /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /src/swi-k8s-opentelemetry-collector /swi-otelcol
COPY --from=wrapper /bin/wrapper /wrapper
COPY --from=journalbinaries / /

ENTRYPOINT ["/wrapper"]
CMD ["/swi-otelcol", "--config=/opt/default-config.yaml"]

