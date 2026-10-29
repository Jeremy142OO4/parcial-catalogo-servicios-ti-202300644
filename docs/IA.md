# Ingeniería asistida por IA en Service Desk

Inicio: 2026-10-27. Actualización: 2026-10-28. Estado vigente: implementación y harness verificados sobre commits publicados. P04 tiene cierre documental en el alcance acordado con Jeremy y salvedad interpretativa, según la [decisión posterior](decisiones/002-interpretacion-cmdb-p04.md). No confundir evidencias del parcial con resultados nuevos.

El índice vigente distingue implementación, revisiones web, loops y commits. Los apartados de planificación archivados más abajo conservan el contexto inicial. El [cierre técnico](evidencias/service-desk-cierre.md) registra las ejecuciones reales de `0b5b8c7` y `461ac08` desde un clon limpio, incluida persistencia, sin confundir una respuesta de IA con una prueba del sistema.

## Índice vigente de cierre, 28 de octubre

Hay nueve registros seleccionados de la ampliación, 008–016. Los mensajes web tienen respuesta y captura reales. Los registros de implementación conservan las solicitudes literales del usuario y no se presentan como chats web ficticios.

| Registro | Uso real | Evidencia |
| --- | --- | --- |
| [008](prompts/008-diseno-service-desk.md) | Diseño, requisitos, relaciones y riesgos, ChatGPT web | Modelo y procesos previos |
| [009](prompts/009-implementacion-backend-service-desk.md) | Migraciones, permisos, CMDB y siete rutas, Codex desktop | Backend PostgreSQL |
| [010](prompts/010-adjuntos-protegidos.md) | Adjuntos, Codex desktop | Tamaño, tipo, ACL y persistencia |
| [011](prompts/011-indicadores-e-interfaz.md) | Indicadores y React, Codex desktop | Snapshots, pausas, portal y consola |
| [012](prompts/012-faq-demostracion-verificacion.md) | FAQ y 37 casos, Codex desktop | Datos separados y verificación final |
| [013](prompts/013-revision-cmdb-cierre.md) | Fuentes CMDB, ChatGPT web | Respuesta inicial con atribución imprecisa |
| [014](prompts/014-revision-cmdb-refinada.md) | Fuente corregida con celdas, ChatGPT web | Atribución corregida y dos pruebas de preservación |
| [015](prompts/015-revision-harness-cierre.md) | Harness inicial, ChatGPT web | Detectó contexto insuficiente de ejecución |
| [016](prompts/016-revision-harness-refinada.md) | Harness con evidencia, ChatGPT web | PASS/SKIP separados y clon limpio pendiente de ejecutar en ese momento |

Las instrucciones operativas siguen rol, objetivo, contexto, entrada, formato y criterio mediante párrafos. No se reemplaza una solicitud literal del usuario por una versión escrita después. El selector web mostró Medium/Media, no identifica por sí solo un modelo exacto. La identificación histórica GPT-5.6 Sol / High no se atribuye a estas sesiones.

### Dos mejoras verificables de prompts

**013 → 014:** el primer mensaje no indicaba C5/C6 y la respuesta trató Hardware/Software como decisiones propias. Se proporcionaron las celdas y se separaron esos valores de las seis subclases complementarias. La segunda respuesta corrigió la atribución y mantuvo P04 parcial. Se añadieron comparación de los 40 valores y comprobación de que el registro de origen no crea CI o relaciones.

**015 → 016:** el primer mensaje describía el runner sin resultados concretos. La revisión pidió evidencia de aislamiento y clon limpio. La segunda instrucción añadió sufijos, semilla, fixture corregido, resultados y diferencia entre SKIP general y persistencia ejecutada por separado. El resultado separó controles disponibles del paso de reproducibilidad pendiente, que se comprueba sobre un SHA real en el [cierre técnico](evidencias/service-desk-cierre.md).

Las respuestas del modelo son propuestas revisadas, no certificados del sistema. Las fuentes originales no se enviaron al chat web.

#### Evidencia visual de las mejoras

Revisión inicial de CMDB, registro 013:

![Inicio de la respuesta inicial sobre las fuentes CMDB](Imagenes/40-prompt-cmdb-inicial.png)

Revisión refinada con las coordenadas de origen, registro 014:

![Inicio de la respuesta corregida sobre las fuentes CMDB](Imagenes/41-prompt-cmdb-refinado.png)

Revisión inicial del harness, registro 015:

![Inicio de la respuesta inicial sobre el harness](Imagenes/42-prompt-harness-inicial.png)

Revisión refinada con resultados y límites, registro 016:

![Inicio de la respuesta refinada sobre el harness](Imagenes/43-prompt-harness-refinado.png)

### Contexto versionado

| Actualización | Motivo | Commit o referencia |
| --- | --- | --- |
| [Alcance y fuentes](contexto/service-desk-alcance-y-fuentes.md) | Separar parcial y ampliación, no inventar taxonomía | `9e3a590` |
| [Modelo y validación](contexto/service-desk-modelo-y-validacion.md) | Revisar propuesta web y distinguir seis tipos de siete procesos | `9e3a590` |
| [Backend y aislamiento](contexto/service-desk-backend-y-aislamiento.md) | Contratos reales, roles y fixtures propios | `a50ee90` |
| [Indicadores e interfaz](contexto/service-desk-indicadores-e-interfaz.md) | Snapshots y corrección del trigger | `0b5b8c7` |
| [FAQ y verificación](contexto/service-desk-faq-y-verificacion.md) | Separar 37 casos de las pruebas | `0b5b8c7` |
| [Cierre y fuentes](contexto/service-desk-cierre-y-fuentes.md) | Origen visible y límites contrastados | `0b5b8c7` |

Las actualizaciones iniciales conservan su fecha y estado histórico. No se transforman en evidencias de funciones que todavía no existían cuando se escribieron. Las diapositivas siguen fuera del repositorio.

### Harness comprobable

[service-desk-test.sh](../scripts/service-desk-test.sh) construye imágenes, comprueba originales durante el build, aplica migraciones, importa catálogo, genera datos y ejecuta pruebas. [service-desk-persistence.sh](../scripts/service-desk-persistence.sh) reinicia base/API y compara registros y un adjunto. [service-desk-demo-verify.sh](../scripts/service-desk-demo-verify.sh) verifica por separado los 37 casos, sin escribir. No se requiere Go o Node instalados en el anfitrión.

Los runners propagan fallos, restringen el proyecto a `-test` y los tests exigen base `_desk_test`. La búsqueda de patrones de claves privadas y tokens es heurística y no certifica ausencia de cualquier secreto. Las pruebas omitidas no cuentan como aprobadas. La evidencia del clon limpio identifica el SHA real, no uno supuesto.

### Loops completos

| Ciclo | Ejecución y evaluación | Corrección y nueva verificación |
| --- | --- | --- |
| Datos y permisos de fixtures | Repetir tests falló al editar grants con tickets pendientes | Colas y CI propios por corrida sin debilitar permisos, repetición aprobada. [Evidencia](evidencias/service-desk-backend.md) |
| Defecto funcional de colas | Editar objetivos falló con `record new has no field role` | Migración 007 con ramas separadas por tabla, edición y regresiones aprobadas. [Fallo y corrección](evidencias/service-desk-indicadores-interfaz.md) |
| Medición P16 | Primer plan obtenido antes de cargar 5,000 dispositivos | Precondición de carga y nuevo plan real con índice. [Evidencia](evidencias/service-desk-final.md) |
| Repetición P14 | Una relación ya existente hizo fallar la restricción única | 34 CI exclusivos por ejecución y repetición aprobada sin limpiar base. [Fallo](evidencias/service-desk-final-fallo-p14.txt) |

Los fallos anteriores no fueron deliberados. Límite operativo: tres intentos ante el mismo bloqueo antes de intervención humana. El usuario decide sobre fuentes ambiguas y sobre el alcance. Se declararon como fixtures deliberados los timestamps controlados, el cambio planificado sin aprobación y el trigger temporal de fallo de auditoría, exclusivamente en la base aislada.

### Grafos y defensa

Árbol de clases sin ciclos, distinto del grafo técnico. `depende_de` y `alojado_en` rechazan ciclos. `conectado_a` es simétrica y admite ciclos. El impacto potencial conserva rutas, relaciona servicios y tiene límites de recorrido, no certifica interrupción efectiva. La decisión posterior de P04 cierra el pendiente de un árbol docente adicional, con salvedad interpretativa y sin inventar contenido de las fuentes.

Consultar el [procedimiento de ejecución](../README.md) y el [cierre](evidencias/service-desk-cierre.md). La entrega en Wedy y la defensa oral requieren intervención del estudiante.


## Estructura de los prompts y antecedentes

Los mensajes operativos integran seis elementos mediante párrafos, sin etiquetas grandes dentro del prompt: rol, objetivo, contexto, entrada, formato y criterio de aceptación. Cada registro conserva herramienta y modelo realmente identificados, fecha, resultados aplicados y comprobaciones. Una solicitud literal del usuario no se reescribe como si hubiera sido otra conversación.

Los [prompts 001–007](prompts/) y los registros originales de [contexto](contexto/) y [decisiones](decisiones/) conservan la evidencia del catálogo anterior. No sustituyen los nueve registros de esta ampliación.

La [corrección de administración](evidencias/service-desk-correcciones-auditoria.md) documenta identidad técnica de clases, auditoría y administración delegada. La [revisión de actores y revocación](evidencias/service-desk-permisos-y-actor.md) añade un control negativo sobre una copia temporal y seis casos concurrentes satisfactorios. Su [actualización de contexto](contexto/service-desk-revocacion-y-usuarios.md) explica el cambio. Las correcciones quedaron integradas en el merge `8a103ac`. El [despliegue posterior](evidencias/service-desk-despliegue-final.md) identifica la versión realmente probada.

Los documentos antiguos conservan su estado histórico. El inventario vigente está en los cinco documentos enlazados desde el README principal, no en las menciones de pendientes de esos registros.
