FROM golang


RUN go get github.com/gorilla/mux

WORKDIR /core

COPY . .
EXPOSE 8080

CMD ["go", "run", "."]