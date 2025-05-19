FROM golang:1.21

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

#RUN go install github.com/air-verse/air@latest

COPY . .

RUN go build -o api ./cmd/api
RUN go build -o worker ./cmd/worker

CMD ["go", "run", "./cmd/api"]
#CMD ["air"]