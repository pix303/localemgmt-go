FROM golang:1.23-alpine AS builder

WORKDIR /workspace
COPY ./api/go.mod ./api/
COPY ./api/go.sum ./api/

WORKDIR /workspace/api
RUN go mod download
RUN go mod verify

WORKDIR /workspace
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o api-app ./api/cmd/web-api/main.go


FROM alpine:latest AS runner
RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /workspace/api-app .

EXPOSE 8080

CMD ["./api-app"]
