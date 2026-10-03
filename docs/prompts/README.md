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

- El archivo original del catálogo se conserva sin modificaciones.
- Los resultados se documentan después de cada prompt.
- Las credenciales no se incluyen dentro de los prompts.
- Las pruebas deben ser reproducibles mediante Docker Compose.
- Los criterios se relacionan con los requisitos de la tarea y con las evidencias del repositorio.
