FROM golang

WORKDIR /go/src/CNOnt-Broker/core
RUN export GO111MODULE=on

COPY . .

RUN go build -o main .

WORKDIR /dist
RUN cp /go/src/CNOnt-Broker/core/main .
RUN cp /go/src/CNOnt-Broker/core/secret/kubeconfig .

RUN mkdir -p ./ontology/assets
RUN cp /go/src/CNOnt-Broker/core/ontology/assets/* ./ontology/assets

EXPOSE 8080

CMD ["/dist/main", "--kubeconfig", "kubeconfig"]
