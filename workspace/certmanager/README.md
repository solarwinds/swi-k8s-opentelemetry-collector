```
helm install cert-manager jetstack/cert-manager --namespace test-namespace --create-namespace --version v1.16.1 --set crds.enabled=true
```