# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copiar go mod files
COPY go.mod ./
RUN go mod download

# Copiar código fuente
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o orgmserver .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copiar binario desde builder
COPY --from=builder /app/orgmserver .

# Crear directorio para logs y estado
RUN mkdir -p /tmp /root/logs

# Exponer puerto si es necesario (aunque este servicio no expone HTTP)
# EXPOSE 8080

CMD ["./orgmserver"]

