FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copiar el binario ya compilado (debes asegurarte que el binario esté junto al Dockerfile)
COPY orgmserver .

# Crear directorio para logs y estado
RUN mkdir -p /tmp /root/logs

# Exponer puerto si es necesario (aunque este servicio no expone HTTP)
# EXPOSE 8080

CMD ["./orgmserver"]
