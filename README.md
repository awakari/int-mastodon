# About

Mastodon-specific integration.

# Other

```shell
kubectl create secret generic int-mastodon-client \
  --from-literal=hosts=host1,host2,... \
  --from-literal=keys=key1,key2,... \
  --from-literal=secrets=secret1,secret2,... \
  --from-literal=tokens=token1,token2,...
```
