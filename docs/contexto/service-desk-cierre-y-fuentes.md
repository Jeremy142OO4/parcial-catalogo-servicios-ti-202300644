# Cierre técnico y contraste de fuentes

Fecha: 28 de octubre de 2026. Petición del usuario: completar los pendientes, guardar y subir el trabajo y preparar la defensa.

Actualización posterior a la consulta: Jeremy aclaró que no había contenido docente adicional si no aparecía en las diapositivas recibidas. La [decisión de P04](../decisiones/002-interpretacion-cmdb-p04.md) cierra el pendiente de un árbol adicional, con salvedad interpretativa. Los hallazgos que siguen conservan el contexto anterior a esa respuesta, no son un bloqueo nuevo ni una exigencia de inventar clases.

## Hallazgos que cambian el contexto

Se releyó el enunciado completo y se revisó la lección 22 recibida en Downloads. No se copiaron ni reprodujeron las diapositivas en el repositorio o su documentación. La correspondencia de campos y relaciones se contrasta con el mapeo existente. No se encontró una enumeración de clases que permita atribuir al docente las seis subclases complementarias del proyecto. P04 conserva esa limitación, no se cambia de parcial a aprobado por haber pasado las pruebas de ciclos.

El enunciado exige atributos técnicos dependientes de clase, no un editor genérico de esquemas de atributos. El servidor ya valida los atributos permitidos y la interfaz permite editarlos. Ese editor genérico se reclasifica como mejora opcional, no como un requisito obligatorio pendiente.

Las dos filas originales tienen referencias que requieren decisión humana antes de convertirse en CI operativos. Se agregó una vista de origen de solo lectura, preservación de los 40 valores, carga repetible de la fuente y pruebas de que la lectura/importación de origen no crea CI o relaciones ficticias. La preparación de demostración ahora también registra esta fuente, sin modificar sus 37 tickets.

## Controles de cierre

El build ejecuta comprobación SHA-256 de los originales y una búsqueda heurística de patrones de claves privadas o tokens en el árbol fuente. No se presenta esa búsqueda como una auditoría completa de secretos. Se añadieron negativas explícitas de P10 y P15. La intervención SQL que prepara un cambio planificado sin aprobación es un fixture deliberadamente inconsistente, exclusivo de la base de pruebas.

Se preparó la [guía de defensa](../../README.md) y se asociará la validación desde clon limpio al SHA real del código, sin inventar commits ni modificar el tag del parcial.
