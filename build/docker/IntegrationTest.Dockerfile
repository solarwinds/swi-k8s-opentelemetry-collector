# syntax=docker/dockerfile:1
FROM python:3.13-alpine3.19

ADD --checksum=sha256:1d7f49f5aa52670d5f20970a9058894c0e82fee9c40dd935187168b8a9d96fa6 \
    https://clientdownload.catonetworks.com/public/certificates/CatoNetworksTrustedRootCA.pem \
    /usr/local/share/ca-certificates/cato.crt
RUN  update-ca-certificates

# gcc musl-dev python3-dev are needed for `clickhouse-connect` python package
RUN apk add --update --no-cache curl ca-certificates bash gcc musl-dev python3-dev

ARG KUBECTL_VERSION=1.25.2
# Install kubectl (same version of aws esk)
RUN curl -sLO https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl && \
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