# Evidencia: ciclo de Docker y persistencia

Fecha: 2026-10-01

## Control ejecutado

```bash
docker compose up --build -d
```

La imagen de la API se construyó correctamente. PostgreSQL 18 falló al iniciar porque el volumen estaba montado en `/var/lib/postgresql/data`, una ruta incompatible con la estructura de datos de las imágenes 18+.

## Corrección

Se cambió el montaje a:

```yaml
volumes:
  - catalogo_pgdata:/var/lib/postgresql
```

Se eliminó únicamente el volumen de prueba creado por este intento y se repitió el control. No se eliminaron datos de evaluación ni el Excel original.

## Segundo control

Se ejecutó nuevamente:

```bash
docker compose up --build -d
```

Resultado: los contenedores `db` y `api` quedaron activos. PostgreSQL reportó estado saludable y la API publicó el puerto `8080`.

Controles funcionales:

```text
/api/health -> {"status":"ok"}
/api/ready  -> {"status":"ready"}
schema_migrations -> 1
service_classes -> 2
companies -> 0
```

También se ejecutó `go test ./...` dentro de un contenedor Go y todos los paquetes finalizaron correctamente. La base quedó preparada para el siguiente ciclo de desarrollo. Todavía no se cargan usuarios ni datos de catálogo en PostgreSQL.

## Persistencia del catálogo

Con los servicios activos se ejecutó el importador con `--database-url`. El resultado fue:

```text
Persistencia PostgreSQL: OK
Importación validada: 12 nivel 1, 46 nivel 2, 4 observaciones, 51 filas omitidas
```

Controles directos en PostgreSQL:

```text
schema_migrations -> 2
services_level1 -> 12
services_level2 -> 46
import_observations -> 4
import_runs con estado succeeded -> 1 en el primer control
```

La operación utiliza `ON CONFLICT` para actualizar por código y un índice único para no duplicar la evidencia de nombres fuente al repetir la importación.

Se repitió el mismo comando de importación para comprobar idempotencia. Los conteos permanecieron en 12 servicios nivel 1 y 46 nivel 2. El único incremento fue el esperado en `import_runs`, que pasó a 2 importaciones exitosas.
