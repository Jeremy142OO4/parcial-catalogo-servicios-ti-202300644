# Diccionario de datos y mapeo del Excel

Todas las claves primarias son `id BIGINT GENERATED ALWAYS AS IDENTITY`. Las relaciones utilizan claves foráneas y no eliminan registros en cascada.

| Tabla | Campos y tipos | Relaciones y restricciones |
|---|---|---|
| companies | code, name TEXT, is_active BOOLEAN | code único |
| areas | company_id BIGINT, code, name TEXT, is_active BOOLEAN | FK companies, único company_id + code |
| departments | area_id BIGINT, code, name TEXT, is_active BOOLEAN | FK areas, único area_id + code |
| sections | department_id BIGINT, code, name TEXT, is_active BOOLEAN | FK departments, único department_id + code |
| positions | section_id BIGINT, code, name TEXT, is_active BOOLEAN | FK sections, único section_id + code |
| app_users | position_id BIGINT, full_name, username, email, password_hash, role TEXT, is_active BOOLEAN | FK positions, username y email únicos, roles admin y consulta |
| sessions | user_id BIGINT, token_hash TEXT, expires_at, created_at TIMESTAMPTZ | FK app_users, token_hash único, expiración de ocho horas |
| service_classes, criticalities, service_types | name TEXT | Nombre único, mantenimiento protegido en servidor |
| services_level1 | code, canonical_name TEXT, is_active, review_required BOOLEAN, source_sheet, source_range TEXT, source_row INTEGER | code único |
| services_level2 | service_level1_id BIGINT, code, name, active_value, description, metric TEXT, minimum, maximum NUMERIC, service_class_id, criticality_id, service_type_id BIGINT, is_active, review_required BOOLEAN, source_sheet, source_range TEXT, source_row INTEGER | FK nivel 1 y opciones, código único, minimum ≤ maximum cuando ambos existen |
| services_level1_source_names | service_level1_id BIGINT, original_name, source_sheet, source_range TEXT, source_row INTEGER, is_canonical BOOLEAN | FK nivel 1, combinación fuente única según migración 002 |
| service_assignments | service_level2_id, section_id, responsible_user_id BIGINT, created_at TIMESTAMPTZ | Un servicio tiene a lo sumo una asignación, responsable opcional de la misma sección |
| import_runs | source_file, source_sha256, status TEXT, started_at, finished_at TIMESTAMPTZ, created_count, updated_count, omitted_count, observed_count INTEGER | Estados running, succeeded, failed. Conteos creados y actualizados calculados en transacción |
| import_observations | import_run_id BIGINT, code, severity, message, sheet_name, source_range TEXT, row_start, row_end INTEGER, details JSONB | FK ejecución de importación, severidad info, warning o error |

`source_sha256` está reservado en el esquema. El archivo original puede verificarse externamente mediante SHA-256. No se afirma que ese campo se complete automáticamente.

| Columna Excel | Destino | Regla |
|---|---|---|
| A COD.N1 | services_level1.code | Texto único |
| B SERVICIO nivel 1 | services_level1.canonical_name y services_level1_source_names.original_name | Primer valor canónico, todos los nombres fuente conservados |
| C COD.N2 | services_level2.code | Texto único, sin convertir SE.12.x a número |
| D SERVICIO nivel 2 | services_level2.name | Celda principal dentro de combinaciones |
| E ACTIVO | services_level2.active_value | S, N o NULL, separado del estado administrativo is_active |
| F CLASE | services_level2.service_class_id | FK service_classes, NULL si ausente |
| G CRITICIDAD | services_level2.criticality_id | FK criticalities, NULL si ausente |
| H TIPO | services_level2.service_type_id | FK service_types, etiquetas originales |
| I Descripción | services_level2.description | Texto opcional, NULL si ausente |
| J Métrica | services_level2.metric | Texto opcional, NULL si ausente |
| K Mínimo | services_level2.minimum | NUMERIC nullable, sin convertir vacío en cero |
| L Máximo | services_level2.maximum | NUMERIC nullable, valida rango |

Las opciones E112:H122 no crean servicios. La organización, usuarios y asignaciones son datos nuevos de demostración, no proceden del Excel. La navegación se pagina en React en grupos de doce después de aplicar los filtros del servidor.
