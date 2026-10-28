# Interpretación de las fuentes CMDB y cierre documental de P04

Fecha: 28 de octubre de 2026. Estado: decisión de alcance acordada con el usuario, con salvedad de interpretación del enunciado.

## Aclaración recibida

Se preguntó a Jeremy si existía un árbol adicional de clases de dispositivos aparte del material recibido. Su respuesta literal fue:

> si no lo viste en las diapositivas no lo enseño

Esta es una aclaración del usuario sobre el material disponible. No se presenta como una confirmación directa del catedrático, una verificación nueva de la grabación o una modificación del enunciado escrito.

## Decisión

No tratar un árbol docente no enumerado como una funcionalidad pendiente de desarrollar. Justificar la CMDB mediante sus campos, tablas, relaciones y valores observables, sin inventar niveles ni denominaciones atribuidos a las fuentes.

Hardware y Software proceden de C5 y C6 de la hoja CMDB. Servidor, Equipo de usuario, Red, Impresión, Almacenamiento y Energía son subclases complementarias del proyecto para organizar dispositivos y validar atributos técnicos. No se presentan como una transcripción docente. La organización Empresa → Área → Departamento → Sección → Puesto → Usuario sigue siendo independiente de esa clasificación.

## Correspondencia y evidencia

| Elemento | Tratamiento | Evidencia existente |
| --- | --- | --- |
| Campos originales | Diccionario de 20 campos y conservación de 40 valores de dos filas | [Modelo y mapeo](../MODELO.md), `CMDB_original_rows_and_readonly_review` |
| Tipos observables | Hardware y Software conservados como valores de origen | Hoja CMDB, C5/C6, vista administrativa Origen del Excel |
| Clasificación complementaria | Seis subclases declaradas como diseño propio, sin atribución docente | Modelo, clases persistidas y generador reproducible |
| Relaciones técnicas | Separadas de la clasificación y de la organización | P14, vínculos CI/servicio y recorrido de impacto potencial |
| Padres e integridad | Rechazo de padres inexistentes y ciclos, incluida carrera A→B/B→A | P04 de integración y prueba concurrente |
| Referencias incompletas | Valores originales conservados, sin CI ni relaciones ficticias | Importación repetible de origen y conteos operativos invariantes |

Las ejecuciones reales están en el [cierre técnico](../evidencias/service-desk-cierre.md), sobre los commits allí identificados. Esta aclaración no cambia código, datos, pruebas ni sus salidas históricas y no requiere simular una nueva ejecución funcional.

## Estado que debe comunicarse

P04 tiene integridad y correspondencia con la información disponible verificadas, con salvedad interpretativa sobre la jerarquía mencionada por el enunciado. El pendiente documental de obtener un árbol adicional queda cerrado para el alcance acordado con Jeremy. No equivale a afirmar que las fuentes enumeran las seis subclases ni a garantizar la aceptación del catedrático.

Las revisiones y respuestas anteriores que calificaban P04 como parcial conservan su valor histórico, pues preceden a esta aclaración. No se alteran prompts, capturas, stdout, SHA o fechas para aparentar que la decisión existía al ejecutar las pruebas.

Si el catedrático proporciona posteriormente una jerarquía concreta o exige otra interpretación, deberá contrastarse con esta decisión y registrarse el cambio. No se anticipa esa información ni se rellena por suposición.
