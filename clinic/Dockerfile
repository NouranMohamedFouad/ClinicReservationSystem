FROM golang:1.17-alpine3.14

WORKDIR /app

COPY main.go .
COPY go.mod .
COPY go.sum .

RUN go mod download

RUN go build -o clinic .

CMD [ "./clinic" ]

EXPOSE 12345