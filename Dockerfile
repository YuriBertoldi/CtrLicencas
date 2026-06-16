FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o ctrllicenca .

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/ctrllicenca .
COPY templates/ templates/
COPY static/ static/
EXPOSE 8081
CMD ["./ctrllicenca"]
