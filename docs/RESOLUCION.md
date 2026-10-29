# Resolución: catálogo, Service Desk y CMDB

Jeremy Estuardo Orellana Aldana · Carné 202300644 · Software Avanzado.

Documento consolidado el 28 de octubre de 2026. Describe lo implementado, no una propuesta pendiente. La base del parcial permanece en `parcial-v2.0`. Service Desk fue integrado en main mediante `8a103ac`. La actualización posterior `7681a47` conserva scripts operativos de la laptop.

## 1. Problema y alcance

El sistema reemplaza la gestión aislada en Excel por un catálogo trazable y un punto de atención para solicitudes. Conserva organización Empresa → Área → Departamento → Sección → Puesto → Usuario, autenticación local, bajas lógicas, asignaciones, catálogos auxiliares y filtros.

La ampliación agrega siete procesos de tickets, colas con permisos independientes, clasificación y relaciones de CI, aprobaciones, comentarios públicos e internos, adjuntos protegidos, FAQ e indicadores. La demo tiene 12 servicios nivel 1, 46 nivel 2, siete colas, 37 tickets y 5,000 dispositivos activos. Los conteos de las bases de pruebas son distintos porque contienen fixtures.

No se implementan facturación, KEDB independiente, correo real obligatorio ni cambios automáticos de permisos en sistemas externos. Accesos controla solicitudes, decisiones y evidencias de provisión o revocación, no administra directorios ajenos.

## 2. Arquitectura y tecnologías

```text
React/Vite → Handler/Controller → Service → Repository → PostgreSQL
```

Monolito modular. React organiza las vistas, Go recibe peticiones y aplica reglas, PostgreSQL conserva integridad y transacciones. Las consultas de Service Desk se encapsulan en su módulo de repositorio y servicios, no en handlers HTTP. Docker Compose ejecuta PostgreSQL, API y Nginx/frontend sin instalar los lenguajes en el anfitrión.

Los siete flujos comparten infraestructura de tickets pero tienen rutas y guardas propias. Sus detalles se almacenan en `tickets.details` JSONB. Estados y procesos tienen restricciones referenciales, mientras las transiciones se validan en Go. No existen las siete tablas de detalles ni la tabla editable de transiciones planteadas en el diseño inicial.

El esquema implementado y los originales están en [MODELO](MODELO.md). Los flujos y contratos están en [PROCESOS](PROCESOS.md). La preparación reproducible está en el [README principal](../README.md).

## 3. Catálogo original y decisiones de importación

Se conserva `data/CatalogoServicios.xlsx` sin modificaciones. La hoja de servicios usa datos A4:L101 y listas E112:H122.

1. Una combinación se resuelve desde su celda principal, solo dentro de su rango.
2. No se crea un servicio por cada fila visual. Las continuaciones y opciones no generan registros.
3. `SE.12` usa `Suministrar Analitica` como nombre canónico y conserva `Mantener Tableros de Control` como evidencia de conflicto con observación.
4. Los códigos `SE.12.1`, `SE.12.2` y `SE.12.3` son texto, sin conversión numérica.
5. Los atributos vacíos permanecen desconocidos y en revisión, sin inventar clase, criticidad, tipo, métrica o estado activo fuente.
6. Los registros conservan hoja, fila, rango y transformaciones. Reimportar no duplica códigos.

Las reglas ejecutables y casos de filas sin código están en la [decisión de importación](decisiones/001-reglas-importacion-catalogo.md). El estado administrativo `is_active` no se confunde con el valor de activo ausente del Excel. El diccionario y la asignación de columnas están en [MODELO](MODELO.md).

## 4. CMDB y generación

La CMDB distingue clasificación, relaciones técnicas y organización. Los ejemplos originales de `CMDB.xlsx` se guardan como origen con celdas y SHA-256. No se crean CI ficticios para resolver `Base de Datos1` u otras referencias ambiguas.

Hardware y Software proceden del Excel. Las seis subclases de Hardware son decisiones propias. La aclaración de Jeremy no equivale a una aprobación directa del catedrático ni transforma esas extensiones en contenido docente. La [decisión P04](decisiones/002-interpretacion-cmdb-p04.md) conserva esa salvedad.

El generador usa dataset, semilla y ordinal para obtener códigos deterministas. Repetir el conjunto no duplica ni borra dispositivos. Las referencias operativas existen y los atributos dependen del perfil técnico. La migración 009 separa ese perfil inmutable del código editable de la clase, de modo que renombrarla no rompe la validación ni la generación.

La clasificación rechaza padres inexistentes y ciclos. Las relaciones `depende_de` y `alojado_en` rechazan ciclos, mientras `conectado_a` es simétrica y los admite. El recorrido inverso devuelve rutas y servicios potencialmente afectados, con límite de profundidad. No certifica una interrupción real.

## 5. Seguridad, permisos y concurrencia

Las contraseñas usan bcrypt con sal. Los tokens de sesión son aleatorios, se almacenan como hash SHA-256, expiran y se invalidan al cerrar sesión. La autorización se evalúa en el servidor por usuario, objeto y acción, no depende de ocultar botones.

Consulta no escribe, incluso combinada con admin, agente o grants. La administración delegada permite editar y conceder permisos de una cola concreta, no crear otras colas ni cambiar roles globales. Aprobar, atender y transferir siguen siendo permisos independientes.

Las acciones críticas usan transacciones, bloqueo y versión del ticket. Dos reclamaciones simultáneas producen un ganador y un conflicto. Las escrituras bloquean las colas antes de consultar grants, para no continuar con un permiso revocado durante la espera. Los métodos administrativos antiguos también revalidan al actor después de bloquearlo y conservan su identidad.

Las bajas rechazan dependencias pendientes en lugar de dejar responsables huérfanos. El historial de tickets, FAQ y cambios administrativos conserva actor, fecha y snapshots en la misma transacción. Las migraciones 009 y 010 cubren relaciones eliminadas, configuración y usuarios, sin incluir hashes de contraseñas. Los cambios anteriores a esas migraciones no se reconstruyen artificialmente.

Adjuntos: máximo 10 MiB, firma y extensión compatibles, nombres seguros y almacenamiento privado. La descarga exige autorización y verifica tamaño y SHA-256. Público significa visible a lectores autorizados del ticket, no acceso anónimo. Los límites detallados están en [PROCESOS](PROCESOS.md).

## 6. IA y decisiones humanas

La IA apoyó análisis, implementación, pruebas y documentación. Las decisiones humanas incluyeron el nombre canónico de SE.12, la separación entre datos de fuente y diseño propio, el alcance de P04, la arquitectura y el aislamiento de pruebas. El estudiante debe comprender y defender el código, no presentar una respuesta del modelo como certificación.

[IA](IA.md) conserva nueve registros de la ampliación, dos mejoras de prompts, más de tres actualizaciones justificadas de contexto y ciclos con resultados reales. Se distinguen fallos accidentales de controles negativos deliberados. La aplicación terminada no consume un modelo ni requiere una cuenta de IA.

## 7. Verificación y recuperación

La [matriz de pruebas](PRUEBAS.md) relaciona P01–P18 con implementación y evidencias. La versión `8a103ac` se construyó desde un clon nuevo. Se ejecutaron siete rutas, pruebas de permisos, importación, generación, adjuntos e indicadores. P18 se ejecutó aparte, con digests y descarga de archivo antes y después del reinicio.

La actualización de la demo conservó 12/46 servicios, 37 tickets y 5,000 dispositivos. Se respaldaron PostgreSQL y archivos y se verificó la restauración en una base separada. Los resultados, imágenes y SHA constan en la [evidencia de despliegue](evidencias/service-desk-despliegue-final.md). No se usa seed para restaurar datos ni se elimina el volumen al reiniciar.

## 8. Limitaciones y entrega externa

- Persiste un falso positivo al editar determinadas dependencias entre CI porque el chequeo puede incluir la relación reemplazada. Es un defecto conocido, no oculto por el PASS de otras pruebas.
- P04 tiene una salvedad interpretativa de las fuentes y de las subclases propias. No se inventó un árbol adicional atribuido al docente.
- Los objetivos usan tiempo calendario, no horas hábiles. Los snapshots históricos desconocidos no se reconstruyen con objetivos actuales.
- La detección de firma de adjuntos no es antivirus. No hay cuotas acumuladas ni limpieza automática de archivos huérfanos tras una caída entre almacenamiento y commit.
- No se afirma cobertura de toda combinación de concurrencia ni defensa frente al propietario de PostgreSQL que deshabilite triggers.
- Confirmar fecha/hora de entrega en Wedy y acceso del catedrático. El resumen de clase aportado por Jeremy reportó jueves 29 frente a miércoles 28 en el PDF. No se toma el viernes 30 como prórroga automática.
- El tag anterior se conserva. Un tag nuevo no se crea ni se mueve como parte de esta consolidación documental.
