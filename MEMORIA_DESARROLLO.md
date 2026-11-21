# Continuidad — correo unificado

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

## Preparación previa

Avance 2026-11-20: copias de ambos repositorios de GitHub preparadas en ~/Code.
No se copiaron secretos, datos ni logs. La versión fuente exacta de la imagen
DNS activa no puede certificarse: el binario no incorpora revisión VCS.

Implementado: orgmserver publica eventos en NOTIFICATION_DIR y orgmdns es el
único emisor de un resumen HTML/texto simple. Incluye inicio, reconexión, cambio
de IP, resultados DNS confirmados y fallos parciales. Sin novedades no hay correo.
Cola JSON atómica persistente, deduplicación de pendientes, reintentos de SMTP,
ventana de agrupación configurable (90 s por defecto; se evalúa por ciclo).
Pruebas de ambos repositorios, go vet y compilación correctos en Go 1.23 Docker.
Sólo DNS y correo simulados; no se enviaron mensajes de prueba ni se alteró DNS.

Estado: código preparado para revisión, activación en producción pendiente.
Los contenedores operativos conservan su imagen y configuración anteriores.
Siguiente acción concreta: seguir UNIFICACION_CORREOS.md, hacer backup nuevo
verificado y activar coordinadamente ambas imágenes con la misma cola compartida.
No confundir publicar el código con desplegarlo. No tocar el dominio de admin.
