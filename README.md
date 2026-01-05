# ORGMServer - Servicio de Monitoreo con Notificaciones SMTP

Servicio en Go que monitorea la conexión a internet, detecta pérdidas de conexión, y envía notificaciones por correo electrónico.

## Características

- Monitoreo continuo de la conexión a internet
- Detección de pérdida de conexión a internet
- Notificaciones por correo SMTP (Gmail) con IP externa
- Healthcheck HTTP opcional a URL configurable
- Persistencia de estado entre reinicios
- Logs de debug configurables
- Nombre de aplicación configurable

## Variables de Entorno

### Requeridas

- `SMTP_HOST` - Servidor SMTP (default: `smtp.gmail.com`)
- `SMTP_PORT` - Puerto SMTP (default: `587`)
- `SMTP_USER` - Email de Gmail para envío
- `SMTP_PASSWORD` - Contraseña de aplicación de Gmail
- `EMAIL_TO` - Email destinatario de las notificaciones

### Opcionales

- `APP_NAME` - Nombre de la aplicación para los correos (default: `ORGMServer`)
- `HEALTHCHECK_URL` - URL para enviar healthchecks HTTP cada minuto (si no se define, no se envía)
- `MONITOR_INTERVAL` - Intervalo de monitoreo en segundos (default: `60`)
- `STATE_FILE_PATH` - Ruta del archivo de estado (default: `/tmp/orgmserver_state.json`)

## Configuración de Gmail

Para usar Gmail como servidor SMTP, necesitas:

1. Habilitar la verificación en 2 pasos en tu cuenta de Google
2. Generar una "Contraseña de aplicación" desde: https://myaccount.google.com/apppasswords
3. Usar esa contraseña de aplicación como `SMTP_PASSWORD`

## Docker Compose

```yaml
version: '3.8'

services:
  orgmserver:
    image: orgmcr.or-gm.com/osmargm1202/orgmserver:latest
    container_name: orgmserver
    restart: unless-stopped
    environment:
      - SMTP_HOST=smtp.gmail.com
      - SMTP_PORT=587
      - SMTP_USER=tu-email@gmail.com
      - SMTP_PASSWORD=tu-app-password-de-gmail
      - EMAIL_TO=destinatario@email.com
      # Opcional: Nombre de la aplicación
      - APP_NAME=ORGMServer
      # Opcional: URL para healthcheck
      - HEALTHCHECK_URL=https://tu-healthcheck-url.com/ping
      # Opcional: Intervalo de monitoreo en segundos
      - MONITOR_INTERVAL=60
      # Opcional: Ruta del archivo de estado
      - STATE_FILE_PATH=/tmp/orgmserver_state.json
    volumes:
      # Persistir estado entre reinicios
      - ./state:/tmp
      # Opcional: Persistir logs
      - ./logs:/root/logs
```

## Build y Push de la Imagen

```bash
# Build
docker build -t orgmcr.or-gm.com/osmargm1202/orgmserver:latest .

# Push
docker push orgmcr.or-gm.com/osmargm1202/orgmserver:latest
```

## Ejecución Local

```bash
# Con debug
go run main.go --debug

# Sin debug
go run main.go
```

## Funcionamiento

1. **Al iniciar**: El servicio envía un correo indicando que el servidor ha sido iniciado, con la IP externa actual. Si se detiene e inicia manualmente, siempre mostrará "iniciado", no detectará pérdida de internet.

2. **Monitoreo continuo**: Cada minuto (configurable), verifica la conexión a internet intentando obtener la IP externa.

3. **Detección de desconexión**: Si se pierde la conexión mientras el servicio está corriendo, guarda el timestamp de la desconexión.

4. **Detección de reconexión**: Cuando se restaura la conexión (mientras el servicio sigue corriendo), calcula la duración de la desconexión y envía un correo de "Conexión restaurada" con el tiempo sin conexión.

5. **Healthcheck opcional**: Si `HEALTHCHECK_URL` está configurado, envía una solicitud HTTP GET cada minuto para indicar que el servicio está funcionando.

## Tipos de Notificaciones

El servicio envía solo 2 tipos de correos:

- **Servidor Iniciado**: Se envía cada vez que el servicio inicia, indicando que está funcionando y activo, junto con la IP externa.

- **Conexión Restaurada**: Se envía cuando se restaura la conexión a internet después de una desconexión detectada, indicando el tiempo que duró la desconexión.

## Logs

Los logs se guardan en `logs/orgmserver.log` cuando se ejecuta con `--debug` o cuando se habilita el modo debug.

## Notas

- El servicio solo detecta pérdidas de conexión a internet mientras está corriendo
- Si detienes e inicias el servicio manualmente, siempre mostrará "iniciado", no detectará pérdida de internet
- El archivo de estado se guarda en `/tmp/orgmserver_state.json` por defecto
- Para producción, monta un volumen persistente para el estado y logs

