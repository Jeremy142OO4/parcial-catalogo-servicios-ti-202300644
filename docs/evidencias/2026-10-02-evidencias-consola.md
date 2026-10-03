# Evidencias de consola: Docker, importación y pruebas

Estas evidencias fueron ejecutadas en la laptop remota `192.168.1.23` el 2 de octubre de 2026. Las imágenes muestran la salida de consola real con fecha y hora local.

## Docker y API

- ![26 — contenedores activos](../Imagenes/26-docker-contenedores.png)
- ![27 — endpoints health y ready](../Imagenes/27-api-health-ready.png)
- ![28 — migraciones aplicadas](../Imagenes/28-migraciones.png)

## Importación e idempotencia

- ![29 — primera importación y conteos](../Imagenes/29-importacion-conteos.png)
- ![30 — segunda importación sin duplicados](../Imagenes/30-reimportacion.png)

La importación produjo 12 servicios de nivel 1, 46 de nivel 2, 4 observaciones y 51 filas omitidas. La segunda ejecución mantuvo 46 códigos distintos de nivel 2.

## Suite de aceptación

- ![31 — pruebas P01–P05](../Imagenes/31-pruebas-p01-p05.png)
- ![32 — pruebas P06–P08](../Imagenes/32-pruebas-p06-p08.png)
- ![33 — pruebas P09–P12](../Imagenes/33-pruebas-p09-p12.png)

Las doce pruebas finalizaron con `PASS`.

## Persistencia del volumen

- ![34 — reinicio y datos conservados](../Imagenes/34-reinicio-volumen.png)

Se detuvieron y levantaron los servicios sin eliminar el volumen. Los conteos permanecieron en 12 nivel 1 y 46 nivel 2. Además, `/api/health` y `/api/ready` respondieron correctamente.
