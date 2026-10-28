# Prompts utilizados

Esta carpeta conserva los mensajes utilizados durante el desarrollo del sistema de gestión del catálogo de servicios de TI.

Cada documento muestra qué se solicitó, qué resultado se aplicó y cómo se comprobó. Los mensajes están escritos de forma directa para que puedan reutilizarse durante una nueva implementación o una revisión del proyecto.

## Organización de cada documento

Los prompts siguen una estructura común, integrada dentro del mensaje para mantener una redacción natural.

| Elemento | Qué define |
| --- | --- |
| Rol | El tipo de conocimiento necesario para resolver la solicitud |
| Objetivo | El resultado que se busca obtener |
| Contexto | Las reglas, restricciones y situación del proyecto |
| Entrada | Los archivos, datos, código o requisitos disponibles |
| Formato | La forma esperada para la respuesta o la implementación |
| Criterio | Las condiciones que permiten aceptar el resultado |

## Índice de prompts

### 009. Implementación del backend Service Desk

[Abrir solicitud real, contexto y resultados](009-implementacion-backend-service-desk.md). Migraciones, autorización, CMDB, generador y siete flujos, con integración aislada real.

### 008. Diseño previo de Service Desk

[Abrir prompt, respuesta seleccionada y captura real](008-diseno-service-desk.md). Primera consulta web de la ampliación, con resumen de requisitos y límites explícitos. No constituye una prueba funcional.

### 001. Diseño del importador y trazabilidad

Define las reglas para celdas combinadas, conflicto `SE.12`, códigos de nivel 2, atributos incompletos, filas de continuación y trazabilidad.

[Abrir prompt 001](001-diseno-importador-trazabilidad.md)

### 002. Esquema y persistencia

Diseña el modelo PostgreSQL, las relaciones, las restricciones y la migración inicial del sistema.

[Abrir prompt 002](002-esquema-y-persistencia.md)

### 003. Docker y persistencia PostgreSQL

Prepara el entorno reproducible con Docker Compose, healthchecks, puertos y volumen persistente.

[Abrir prompt 003](003-docker-persistencia.md)

### 004. Persistencia del catálogo importado

Conecta el importador con PostgreSQL mediante la arquitectura Service y Repository y conserva la trazabilidad.

[Abrir prompt 004](004-persistencia-importacion.md)

### 005. Autenticación, organización e interfaz

Implementa autenticación, roles, organización, usuarios, asignaciones e interfaz React.

[Abrir prompt 005](005-auth-organizacion-interfaz.md)

### 006. CRUD, filtros y pruebas de aceptación

Completa el mantenimiento del catálogo, la búsqueda, los filtros y la suite de pruebas P01 a P12.

[Abrir prompt 006](006-crud-filtros-pruebas.md)

## Convenciones

### 013–016. Revisiones reales del cierre

- [013, revisión CMDB inicial](013-revision-cmdb-cierre.md).
- [014, corrección de atribución y pruebas de origen](014-revision-cmdb-refinada.md).
- [015, revisión del harness inicial](015-revision-harness-cierre.md).
- [016, revisión con evidencia y límites](016-revision-harness-refinada.md).

Son cuatro envíos reales a ChatGPT web, con capturas. Los pares 013→014 y 015→016 documentan las mejoras sin inventar conversaciones.

### 012. FAQ, demostración y verificación final

[Abrir solicitud, estructura y resultados reales](012-faq-demostracion-verificacion.md). Incluye aislamiento de la demostración, 37 tickets, ocho artículos y evidencias de P01–P18, con limitaciones explícitas.

### 011. Indicadores e interfaz de Service Desk

[Abrir solicitud, contexto y evaluación](011-indicadores-e-interfaz.md). Registra la implementación, corrección del trigger y pruebas reales, sin reconstruir un chat ficticio.

### 010. Adjuntos protegidos de tickets

[Abrir solicitud, decisiones y resultado](010-adjuntos-protegidos.md). Incluye el registro real de implementación y enlaces a pruebas de carga, tamaño, tipo y descarga autorizada.

### 007. Iteraciones del cierre de auditoría

[Abrir las dos mejoras documentadas](007-iteraciones-cierre-auditoria.md). Distingue las solicitudes literales del usuario de las instrucciones operativas formuladas por el asistente y registra los problemas observados.

- El archivo original del catálogo se conserva sin modificaciones.
- Los resultados se documentan después de cada prompt.
- Las credenciales no se incluyen dentro de los prompts.
- Las pruebas deben ser reproducibles mediante Docker Compose.
- Los criterios se relacionan con los requisitos de la tarea y con las evidencias del repositorio.
