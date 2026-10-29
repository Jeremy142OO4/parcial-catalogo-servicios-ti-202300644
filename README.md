# Service Desk y catálogo de servicios de TI

Software Avanzado · Jeremy Estuardo Orellana Aldana · Carné 202300644.

Aplicación individual en Go, React/Vite y PostgreSQL. Conserva el catálogo del parcial e incorpora siete procesos de tickets, CMDB, permisos por cola, aprobaciones, adjuntos, FAQ e indicadores. No necesita suscripción de IA ni claves de modelos para funcionar.

## Documentos de la entrega

| Documento | Contenido solicitado |
| --- | --- |
| [Resolución](docs/RESOLUCION.md) | Alcance, arquitectura, decisiones, seguridad y limitaciones |
| [Modelo](docs/MODELO.md) | Diagrama, diccionario, jerarquía CMDB, glosario y mapeo de fuentes |
| [Procesos](docs/PROCESOS.md) | Siete flujos, tipos, colas, permisos, aprobaciones e indicadores |
| [Ingeniería con IA](docs/IA.md) | Prompts reales, contexto versionado, harness, mejoras y loops |
| [Pruebas](docs/PRUEBAS.md) | Matriz P01–P18, comandos, resultados y evidencias |

Las carpetas `docs/prompts`, `docs/contexto`, `docs/decisiones`, `docs/evidencias` y `docs/Imagenes` son el respaldo de esos documentos, no guías de instalación adicionales.

## Requisitos y arranque desde un clon limpio

Se necesita Git y Docker con Compose. No es necesario instalar Go, Node ni PostgreSQL en el anfitrión. Los scripts de pruebas usan Bash, disponible en Linux o Git Bash. La laptop de desarrollo utiliza Podman con proveedor compatible de Compose.

```bash
git clone https://github.com/Jeremy142OO4/parcial-catalogo-servicios-ti-202300644.git
cd parcial-catalogo-servicios-ti-202300644
docker compose up --build -d
docker compose ps
```

Las migraciones 001–010 se aplican automáticamente al iniciar la API. Esperar su disponibilidad antes de ejecutar seed, evitando dos migradores simultáneos:

```bash
curl --fail http://localhost:8080/api/health
curl --fail http://localhost:8080/api/ready
```

| Acceso | URL |
| --- | --- |
| Interfaz normal | [localhost:5173](http://localhost:5173/) |
| API | [localhost:8080/api/health](http://localhost:8080/api/health) |
| Disponibilidad de base y API | [localhost:8080/api/ready](http://localhost:8080/api/ready) |

El frontend usa un proxy `/api`, sin fijar la IP del equipo en el navegador. PostgreSQL no publica un puerto al anfitrión en este Compose.

### Catálogo, organización y cuentas iniciales

Solo en una base nueva, proporcionar contraseñas locales por variables de entorno:

```bash
read -rs -p 'Contraseña local del administrador: ' SEED_ADMIN_PASSWORD
echo
read -rs -p 'Contraseña local de consulta: ' SEED_CONSULTA_PASSWORD
echo
export SEED_ADMIN_PASSWORD SEED_CONSULTA_PASSWORD
docker compose exec -e SEED_ADMIN_PASSWORD -e SEED_CONSULTA_PASSWORD api ./catalogo-seed
```

El seed importa el Excel original, conserva 12 códigos nivel 1 y 46 nivel 2 y crea organización, `admin-demo`, `consulta-demo` y tres asignaciones. Cada cuenta utiliza la contraseña introducida para su variable. Puede iniciarse sesión con usuario o correo local. No repetir seed para restaurar una demo ya modificada.

Para generar un conjunto CMDB independiente después del seed:

```bash
docker compose exec api ./cmdb-generator --dataset cmdb-demo --seed 202300644 --count 5000
```

Repetir dataset y semilla no duplica códigos. La generación no borra dispositivos ni sustituye ediciones manuales. Los ejemplos originales de CMDB se conservan en un registro de origen, sin resolver automáticamente sus referencias ambiguas.

## Demo completa: seis cuentas y 37 tickets

Este entorno está separado del Compose normal y de las pruebas. En una base nueva:

```bash
docker compose -p catalogo-service-desk-demo -f compose.desk-demo.yaml up --build -d
curl --fail http://localhost:18090/api/ready
read -rs -p 'Contraseña sintética de al menos 12 caracteres: ' DESK_DEMO_PASSWORD
echo
export DESK_DEMO_PASSWORD
docker compose -p catalogo-service-desk-demo -f compose.desk-demo.yaml exec -T -e SEED_ADMIN_PASSWORD="$DESK_DEMO_PASSWORD" -e SEED_CONSULTA_PASSWORD="$DESK_DEMO_PASSWORD" api ./catalogo-seed
docker compose -p catalogo-service-desk-demo -f compose.desk-demo.yaml exec -T -e DESK_DEMO_ALLOW_WRITE=1 -e DESK_DEMO_PASSWORD api ./desk-demo
bash scripts/service-desk-demo-verify.sh
```

Si la consulta `ready` falla durante el arranque, esperar y repetirla antes de continuar. Los comandos de preparación no deben ejecutarse simultáneamente.

| Cuenta | Rol | Contraseña al preparar una base nueva |
| --- | --- | --- |
| `desk-admin` | Administrador | Valor introducido en `DESK_DEMO_PASSWORD` |
| `desk-solicitante` | Solicitante | El mismo valor |
| `desk-agente1` | Agente | El mismo valor |
| `desk-agente2` | Segundo agente | El mismo valor |
| `desk-aprobador` | Aprobador | El mismo valor |
| `desk-consulta` | Solo lectura | El mismo valor |

La contraseña de la instalación existente se entrega por separado, no se publica en Git. La preparación rechaza cuentas existentes con credenciales o roles incompatibles, no los restablece silenciosamente.

La demo contiene siete colas, cinco tickets por proceso salvo incidentes con siete casos I0–I6, ocho artículos FAQ y 5,000 dispositivos. Dataset de dispositivos `demostracion-5000`, semilla `20261028`. Repetir `desk-demo` conserva los casos ya preparados y sus estados, no equivale a restaurar una demo alterada.

Acceso: [localhost:18090](http://localhost:18090/) en la laptop o [192.168.1.23:18090](http://192.168.1.23:18090/) desde otro equipo de su red. La versión verificada y desplegada es `8a103ac`, con [respaldo, reinicio y comprobaciones de navegador](docs/evidencias/service-desk-despliegue-final.md). Esa evidencia no garantiza que un contenedor permanezca encendido después de detenerlo manualmente.

## Pruebas y harness

Ejecutar los runners aislados, nunca la suite con escrituras sobre la demo preparada:

```bash
bash scripts/service-desk-test.sh
DESK_TEST_PROJECT=catalogo-service-desk-test bash scripts/service-desk-persistence.sh
bash scripts/service-desk-demo-verify.sh
```

El primero construye, prepara datos y ejecuta la suite en base `_desk_test`, puertos 18080/18081 y volúmenes propios. El segundo reinicia esa misma base y API, compara digests y descarga un adjunto protegido. El tercero solo consulta la demo de 37 tickets. Puertos alternativos, pruebas históricas del catálogo y resultados están en [PRUEBAS](docs/PRUEBAS.md).

## Logs, apagado y recuperación

Para el entorno normal:

```bash
docker compose logs --tail=100 api
docker compose stop
docker compose start
```

Para la demo:

```bash
docker compose -p catalogo-service-desk-demo -f compose.desk-demo.yaml logs --tail=100 api
docker compose -p catalogo-service-desk-demo -f compose.desk-demo.yaml stop
docker compose -p catalogo-service-desk-demo -f compose.desk-demo.yaml start
```

`stop/start` conserva contenedores y datos. Recrearlos conserva los volúmenes nombrados. No usar `down -v` sobre una instalación que deba conservarse. Para recuperar se necesitan tanto el respaldo PostgreSQL como el volumen de adjuntos. La última [restauración aislada del respaldo](docs/evidencias/service-desk-despliegue-metadatos.txt) fue comprobada sin reemplazar la demo.

Los archivos `release-demo-*.sh`, `release-demo-snapshot.sql` y `deployment-backup-path.txt` registran una actualización concreta de la laptop. Contienen rutas e imágenes de esa instalación. No son el procedimiento general de arranque ni deben ejecutarse a ciegas desde un clon nuevo. El script de actualización está restringido a su SHA original.

## Originales, seguridad y estado de entrega

- Catálogo sin modificar: `data/CatalogoServicios.xlsx`.
- Enunciado y CMDB sin modificar: `data/fuentes-service-desk/`.
- No se incluyen presentaciones docentes, por indicación del usuario. Las diferencias y límites de fuente están en [MODELO](docs/MODELO.md).
- No publicar `.env`, contraseñas, dumps de base de datos ni respaldos de adjuntos. El entorno usa credenciales de infraestructura solo para desarrollo local y no debe exponerse a Internet.
- `parcial-v2.0` conserva la entrega anterior. La ampliación fue integrada en `main`, no se mueve esa etiqueta ni se crea otra automáticamente.
- Existe un pendiente conocido al editar determinadas dependencias entre CI. Está declarado en [RESOLUCION](docs/RESOLUCION.md), aunque las pruebas automatizadas ejecutadas pasen.
- Confirmar personalmente acceso del catedrático, entrega en Wedy y fecha/hora vigentes. Una evidencia técnica no demuestra esas gestiones externas.
