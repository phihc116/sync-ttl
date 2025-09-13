FROM golang:1.24 AS builder
 
WORKDIR /app
 
COPY go.mod go.sum ./
RUN go mod download
 
COPY . .
 
RUN CGO_ENABLED=0 GOOS=linux go build -o updater ./main.go
 
FROM gcr.io/distroless/base-debian12

WORKDIR /app
 
COPY --from=builder /app/updater /app/updater
 
ENTRYPOINT ["/app/updater"]
