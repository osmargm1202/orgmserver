# Continuidad — correo unificado

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
