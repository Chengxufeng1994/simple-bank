#  Build Stage
FROM golang:1.19-alpine3.16 AS builder

WORKDIR /usr/app

COPY . .
RUN go build -o main main.go

# Run stage
FROM alpine:3.16

WORKDIR /usr/app

COPY --from=builder /usr/app/main .
COPY app.env .
COPY start.sh .
COPY wait-for.sh .
COPY db/migrations ./db/migrations

EXPOSE 8080

ENTRYPOINT ["/usr/app/start.sh"]
CMD ["/usr/app/main"]