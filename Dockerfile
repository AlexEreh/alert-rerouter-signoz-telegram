FROM golang:1.25-alpine3.22 AS builder
WORKDIR /app
COPY . .
RUN GOEXPERIMENT=rangefunc CGOENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .

FROM alpine:3.22
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/main .

ENTRYPOINT ["./main"]
