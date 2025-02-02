FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY ./api/go.mod ./
COPY ./api/go.sum ./
RUN ./api/go mod download

COPY /api .

RUN GOOS=linux go build -o api .

FROM alpine:latest  
RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/api .

EXPOSE 8080

CMD ["./api"]
