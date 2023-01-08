#  Build Stage
FROM golang:1.19-alpine3.16 AS builder

WORKDIR /usr/app

COPY . .
RUN go build -o main main.go
RUN apk add curl
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.15.2/migrate.linux-amd64.tar.gz | tar xvz

# Run stage
FROM alpine:3.16

WORKDIR /usr/app

COPY --from=builder /usr/app/main .
COPY --from=builder /usr/app/migrate .
COPY app.env .
COPY start.sh .
COPY wait-for.sh .
COPY db/migrations /migrations

EXPOSE 8080

ENTRYPOINT ["/usr/app/start.sh"]
CMD ["/usr/app/main"]