FROM golang

WORKDIR /go/src/CNOnt-Broker/core
RUN go get github.com/gin-gonic/gin

COPY . .

RUN go build -o main .

WORKDIR /dist
RUN cp /go/src/CNOnt-Broker/core/main .

EXPOSE 8080

CMD ["/dist/main"]
