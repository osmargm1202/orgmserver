# Unificación de avisos de servidor y Cloudflare

Código preparado en los repositorios `orgmdns` y `orgmserver` y activado en
producción el 2026-11-21. Las copias de GitHub conservan sus remotos;
no contienen los `.env`, estado ni logs operativos. La imagen DNS activa tiene
el mismo módulo y dependencias, pero no incluye revisión VCS y su binario difiere
del binario versionado: no se puede certificar identidad de fuentes por checksum.

## Activación coordinada

1. Crear backup nuevo de configuración/estado de ambos servicios y registrar las
   imágenes activas por digest. Cumplir además el backup PostgreSQL con checksums
   y restauración aislada requerido por las instrucciones del servidor antes de
   cualquier cambio de producción. No usar `latest` para esta entrega.
2. Compilar una imagen de cada copia con su Dockerfile, etiquetada con el commit
   publicado. No ejecutar los binarios de la copia con credenciales operativas
   durante las pruebas. No iniciar una segunda instancia del actualizador DNS.
3. Crear una carpeta privada de cola en el host, por ejemplo
   `/home/osmarg/notification-events`, UID/GID `65532:65532`, modo `0770`.
   El runtime DNS usa ese UID y el runtime actual de servidor usa root.
4. Montar esa misma carpeta como `/notifications` en ambos servicios y configurar
   `NOTIFICATION_DIR=/notifications` en ambos. En DNS añadir
   `NOTIFICATION_SETTLE_SECONDS=90`. Conservar las variables SMTP existentes;
   sólo DNS las utiliza en este modo. No montar `.env` del otro servicio.
5. Detener ambos servicios y recrearlos coordinadamente con sus imágenes nuevas y
   los overrides siguientes, conservando sus volúmenes, env_file y comandos actuales.
   Arrancar primero servidor y después DNS. Son overrides para los Compose operativos,
   no sustitutos del despliegue completo.

```yaml
# Override orgmserver (adaptar image al commit real)
services:
  orgmserver:
    image: orgmserver:correo-COMMIT_SERVER
    environment:
      NOTIFICATION_DIR: /notifications
    volumes:
      - /home/osmarg/notification-events:/notifications
```

```yaml
# Override orgmdns (adaptar image al commit real)
services:
  orgmdns:
    image: orgmdns:correo-COMMIT_DNS
    environment:
      NOTIFICATION_DIR: /notifications
      NOTIFICATION_SETTLE_SECONDS: "90"
    volumes:
      - /home/osmarg/notification-events:/notifications
```

6. Comprobar permisos desde ambos runtimes y observar un arranque conjunto: los
   JSON aparecen en la cola; después del período de agrupación DNS envía un solo
   resumen y elimina únicamente los eventos incluidos. No alterar registros DNS
   ni interrumpir la conexión de producción para probar fallos; usar los tests.
   Los eventos que lleguen tras cerrar el resumen quedan para el siguiente ciclo.

## Validación aislada

En cada repositorio: `go test ./...`, `go vet ./...` y `go build` del main
(`./cmd/orgmdns` para DNS; `.` para servidor). Las pruebas usan registros DNS,
SMTP y eventos sintéticos: múltiples cambios, error parcial, reconexión, falta de
IP, ciclo sin novedades, reintentos, persistencia tras reinicio, ventana de
agrupación, publicación atómica, deduplicación y MIME/escape HTML. El servidor en
modo compartido funciona sin conectarse a SMTP. No envían correos reales.

## Límites y reversión

El resumen no se envía mientras continúen llegando eventos dentro de la ventana.
Un archivo JSON corrupto impide vaciar la cola: revisar logs, conservar una copia
y reparar/retirar sólo ese archivo. No borrar la cola completa para desatascarla.
Errores de lectura/borrado se registran. Si falla el almacenamiento, el fallback
en memoria no sobrevive a un reinicio. Un fallo SMTP conserva los archivos.

Para revertir, detener ambos servicios, respaldar la cola pendiente y restaurar
las imágenes/configuraciones anteriores coordinadamente. Los originales volverán
a emitir correos independientes; no dejar servidor antiguo y DNS unificado
activos esperando que se agrupen sus avisos. No toca API, MCP, PostgreSQL, MinIO,
Nginx, Tailscale ni el dominio de la aplicación.

## Producción activada — 2026-11-21

- Activos `orgmdns:correo-9ace318` y `orgmserver:correo-8a48c06`.
  Se corrigió el COPY --chmod del Dockerfile de servidor para el builder instalado.
- Override operativo automático `docker-compose.override.yml` en cada carpeta
  de despliegue, `pull_policy: never` y cola común `/home/osmarg/notification-events`
  montada como `/notifications`. UID/GID 65532:65532, modo 0770.
- Conservados env_file, variables previas, montajes, comandos y restart always.
  DNS mantiene SLEEP_TIME=1 y la ventana de agrupación es de 90 segundos.
- Primera aceptación real: 2026-11-21 12:03:20 UTC, SMTP aceptó un solo resumen con
  los dos eventos de arranque (servidor y DNS). Cola vacía, cero reinicios y cero
  envíos individuales de servidor. No se alteraron registros para provocar pruebas.
- Backup PostgreSQL `20261121T115759Z-acceso-privado`: checksum, restauración aislada
  y huellas coincidentes. Backup específico de configuración/estado e imágenes
  `~/backups/orgm-notifications/20261121T115816Z-unificado/`, con checksums,
  runtime anterior/activo y `rollback.py` privado listo para reversión coordinada.
- Los contenedores ajenos conservan sus IDs; no se cambió el dominio de admin,
  el backend, MCP, Nginx, MinIO ni datos fiscales.

Siguiente acción: observar el siguiente cambio natural de IP/reconexión. No
simular caídas ni editar DNS de producción para probar. Para reconstruir,
compilar ambas versiones y usar los overrides de `deploy/` con los Compose
operativos, no el Compose genérico del repositorio. La cola es persistente;
una caída entre aceptación SMTP y borrado puede repetir el correo.
