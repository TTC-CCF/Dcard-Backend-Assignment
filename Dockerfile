FROM golang:1.20.2-alpine

WORKDIR /app

COPY src/go.mod .
COPY src/go.sum .

RUN go mod download

COPY src .

RUN go build -o main .

CMD ["/app/main"]