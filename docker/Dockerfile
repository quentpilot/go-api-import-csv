FROM golang:1.21

RUN apt-get update && apt-get install -y netcat-openbsd

RUN go install github.com/cespare/reflex@latest

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o api ./cmd/api
RUN go build -o worker ./cmd/worker

CMD ["go", "run", "./cmd/api"]