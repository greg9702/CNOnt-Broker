FROM golang


RUN go get github.com/gin-gonic/gin

WORKDIR /app

COPY . .
EXPOSE 8080

CMD ["go", "run", "."]
