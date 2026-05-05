FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
COPY vendor/ vendor/

COPY cmd/ cmd/
COPY internal/ internal/

RUN go build -mod=vendor -o server ./cmd/server/main.go

FROM alpine:3.19

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/server .
COPY docs/ docs/

RUN mkdir -p uploads

EXPOSE 8080

CMD ["./server"]
