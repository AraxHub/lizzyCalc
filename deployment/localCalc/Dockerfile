FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /calculator ./cmd/calculator

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /calculator /calculator
EXPOSE 8080
CMD ["/calculator"]
