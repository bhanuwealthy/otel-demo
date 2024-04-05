<style>
  .highlight {
  background-color: #f4f4f4;
  padding: 5px;
  border-radius: 5px 5px 15px 15px;
  font-family: "Google Sans";
  border-top: 15px solid black;
  box-shadow: 0 4px 8px 0 rgba(0,0,0,0.2);
}
.md-typeset blockquote  {
  border-left-color: orange !important
}
</style>

# Available in 2 modes
- Docker compose mode
- KinD   cluster mode


## 1. Docker compose mode
```shell
$ make otel
```

## 2. Kubernetes mode

> NOTE: While experimenting this
> 
> Do not use orbStack, it may be stop working due to resource demand and high network activity.
> 
> Instead use  Docker-desktop

### Transform `docker-compose -> deployments`
```shell
$ mkdir kind-deployments
$ cd kind-deployments
$ MONGO_ROOT=dummy \
  kompose --file ../local-setup/local-dockercompose.yaml \
  convert --profile all
```
### KinD
```yaml
$ cat local-setup/local-cluster.yaml

kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
networking:
  disableDefaultCNI: false   # <--- This is for custom CNI like cilium
  podSubnet: 192.168.0.0/16
nodes:
- role: control-plane
- role: worker
- role: worker
```

### Local alias
```shell
alias ktl="kubectl"
alias kctx="kubectx"
```


### Create a  cluster
```
$ kind create cluster \
  --config local-setup/local-cluster.yaml \
  --name otel-demo-cluster

#Confirm
> kctx
kind-otel-demo-cluster <---------
wealthy
```


### Deploy
```
ktl create ns otel
ktl -n otel apply -f kind-deployments
```

### Port forward for local-access
```
$ ktl get svc -n otel

$ ktl -n otel port-forward svc/grafana 3000:grafana-ui > /dev/null &
$ ktl -n otel port-forward svc/otelcol 4317:otel-col > /dev/null &

# kill
$ lsof -i:3000 -t | xargs kill -9
$ lsof -i:4317 -t | xargs kill -9
```

### Standalone services
```shell
# python-app
$ cd py-project
$ source venv-otel-demo/bin/activate

$ make deps run
# running server on http://localhost:8000/docs/
$ deactivate


# ------
# Go binary
$ cd go-project  # From project root
$ make run
# running server on http://localhost:8080

# -----
# Try the apis
$ curl -s http://localhost:8000/ping/ | jq
$ curl -s http://localhost:8000/propagate/ | jq

$ curl -s http://localhost:8080/api/ping/ | jq

```

### Auto instrumentation
```
# Load generator
$ ktl apply -f https://raw.githubusercontent.com/keyval-dev/simple-demo/main/kubernetes/deployment.yaml

# Install auto-instrumentor
$ brew install keyval-dev/homebrew-odigos-cli/odigos
$ odigos install

$ odigos ui

# Then configure the destination as OTLP gRPC with endpoint: otelcol:4317
# Close odigos
```

### Check traces
```
OpenLens ->  GrafanaPod -> Auto open portForward
``` 

### Cleanup 
```
$ ktl delete ns otel

$ kind delete cluster --name otel-demo-cluster
```

