# syntax=docker/dockerfile:1
FROM python:3.13-alpine3.23

ADD --checksum=sha256:1d7f49f5aa52670d5f20970a9058894c0e82fee9c40dd935187168b8a9d96fa6 \
    https://clientdownload.catonetworks.com/public/certificates/CatoNetworksTrustedRootCA.pem \
    /usr/local/share/ca-certificates/cato.crt
RUN  update-ca-certificates

# gcc musl-dev python3-dev are needed for `clickhouse-connect` python package
RUN apk add --update --no-cache curl ca-certificates bash gcc musl-dev python3-dev

ARG KUBECTL_VERSION=1.36.1
RUN ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/') && \
    curl -sLO https://dl.k8s.io/release/v${KUBECTL_VERSION}/bin/linux/${ARCH}/kubectl && \
    mv kubectl /usr/bin/kubectl && \
    chmod +x /usr/bin/kubectl

ARG CI
ENV CI=$CI

WORKDIR /app
COPY /tests/integration/requirements.txt /integration/requirements.txt
RUN ls /integration
RUN pip install --no-cache-dir --upgrade -r /integration/requirements.txt
COPY /tests/integration/ .

CMD ["pytest", "-s", "--tb=short"]