# Collector Operator 



## Required
- Collector operator is generated via operator-sdk
```shell
sudo apt-get update
sudo apt-get install -y curl tar
curl -LO https://github.com/operator-framework/operator-sdk/releases/download/v1.39.0/operator-sdk_linux_amd64
chmod +x operator-sdk_linux_amd64
sudo mv operator-sdk_linux_amd64 /usr/local/bin/operator-sdk
```

## Generate operator image and operator bundle

To generate operator image and bundle image run 

```shell
./.github/operator.sh
```
images generate following images
- solarwinds/solarwinds-otel-operator:$VERSION
- solarwinds/solarwinds-otel-operator:$VERSION-bundle

## Deploy operator via OLM 

```shell
operator-sdk olm install
operator-sdk run bundle solarwinds/solarwinds-otel-operator:0.0.0-bundle
```

## Deploy operator manually
In generated directory `swi-otel-collector`
```shell
make deploy
```

