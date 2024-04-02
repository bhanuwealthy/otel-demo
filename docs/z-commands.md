


## Transform `docker-compose -> deployments`
```shell
> MONGO_ROOT=dummy \
  kompose --file ../local-setup/local-dockercompose.yaml \
  convert --profile all
```
## KinD
```yaml
> cat local-setup/local-cluster.yaml

kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
networking:
  disableDefaultCNI: true   # <--- This is for custom CNI like cilium
  podSubnet: 192.168.0.0/16
nodes:
- role: control-plane
- role: worker
- role: worker
```

## Local alias
```shell
alias ktl="kubectl"
alias kctx="kubectx"
```


## Create a  cluster
```
> kind create cluster \
  --config local-setup/local-cluster.yaml \
  --name otel-demo-cluster

#Confirm
> kctx
kind-local-cilium-cluster
wealthy
```


## Deploy
```
ktl create ns otel
ktl -n otel apply -f kind-deployments
```

## Port forward for local-access
```
$ ktl get svc -n otel

$ ktl -n otel port-forward svc/grafana 3000:grafana-ui &
$ ktl -n otel port-forward svc/otelcol 4317:otel-col &

# kill
$ lsof -i:3000 -t | xargs kill -9
$ lsof -i:4317 -t | xargs kill -9
```

## Cleanup 
```
$ ktl delete ns otel
```


