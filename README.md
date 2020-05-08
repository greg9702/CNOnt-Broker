# __Overview__
CNOnt-Broker is a system which provides connection between cloud native ontology <br>
and Kubernetes cluster. Using Web Ontology Language (OWL) language. <br>
Application expose web API which is used by a client application. <br>

![Image description](docs/assets/system_overview.png)

### __v1.0 version features__
- visualize ontology
- apply ontology to a Kubernetes cluster
- create ontology based on a Kubernetes cluster

### __How to run__
```
git clone https://github.com/greg9702/CNOnt-Broker.git
cd CNOnt-Broker
docker-compose up --build
```

---

### __v1.0 ROADMAP__

__v0.1__
- [x] Add docker-compose
- [ ] Communicate client and core application
- [ ] Elaborate concept of the "top" ontology


__v0.2__
- [ ] Design and implement API
- [ ] Add config file

__v0.3__
- [ ] Add Parser to the project

__v0.4__
- [ ] Finish client application
