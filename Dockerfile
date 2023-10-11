FROM golang:1.21

ENV GIN_MODE=release

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o ./app

EXPOSE 8080

CMD ["./app"]
