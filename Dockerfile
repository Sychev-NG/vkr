FROM golang:1.25-alpine

WORKDIR /usr/src/app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV PORT=8080
EXPOSE 8080

CMD ["go", "run", "cmd/api/main.go"]
