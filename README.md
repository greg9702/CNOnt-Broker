# __Overview__
CNOnt-Broker is a system which provides connection between cloud native ontology <br>
and Kubernetes cluster. Using Web Ontology Language (OWL) language. <br>
Application expose web API which is used by a client application. <br>

![Image description](docs/assets/system_overview.png)

---

### __How to run__

```
git clone https://github.com/greg9702/CNOnt-Broker.git
cd CNOnt-Broker
docker-compose up --build
```

If you would like to run server directly on your machine run:
```
export GO111MODULE=on
cd core
go run main.go --kubeconfig <PATH_TO_KUBE_CONFIG> --logLevel <LOGLEVEL>
```
> Make sure to set _GO111MODULE_ to on. Without this issues with dependencies can occur.


#### __Cluster setup__
Install [kind](https://github.com/kubernetes-sigs/kind) -  tool for running local Kubernetes clusters using Docker container "nodes". <br>
If you have go (1.11+) and docker installed:
```
GO111MODULE="on" go get sigs.k8s.io/kind@v0.8.1
```
If you would like to have access to kind from your console run:
```
kubectl config use-context kind-<cluster name>
```
Set up cluster by running:
```
cd cluster
./setup.sh
```
You are ready to go!

File _cluster-config.yaml_ contains configuration for cluster. Visit [link](https://github.com/kubernetes-sigs/kind) for more details.

Script by default creates admin account, which access token can be obtained by running script _getadmintoken.sh_ in _cluster_ directory.

#### Kuberentes client API documentation

Documentation can be found [here](https://godoc.org/k8s.io/client-go/kubernetes).

#### Issues

According to https://github.com/kubernetes/kubeadm/issues/1292, there can occur bug where _corde-dns_ remains in _CrashLoopBackOff_ state.<br>
To fix this:
```
kubectl -n kube-system edit configmap coredns
```
Remove or comment out the line with loop, save and exit.
```
kubectl -n kube-system delete pod -l k8s-app=kube-dns
```

---

### __System functionality and architecture__

System is able to create deployment based on the ontology file.

Every system element runs in its own docker container. <br> There are three of them:
- `core` - Web API server
- `cluster` - Kubernetes cluster
- `client` - client application created in React

Kubernetes cluster creates proxy on port `8001` and can be accessed from other containers and host machine. Client application exposes port `3000` on `localhost`. Server application exposes port `8080`.

Deployment is created based on file `core/ontology/asssets/CNOnt.owl` which is OWL file with functional syntax. If file is incorrect in some way, deployment won't be created and API would return error code and message.

Server exposes three enpoints:
- `api/v1/create-deployment` - creates deployment
- `api/v1/delete-deployment` - delete deployment if exists
- `api/v1/preview-deployment` - returns preview of deployment
- `/api/v1/serialize-cluster-conf` - create a cluster mapping based on used ontology template file
