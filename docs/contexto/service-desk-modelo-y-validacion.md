# Contexto de modelo y validación de Service Desk

Fecha: 2026-10-27. Segunda actualización de contexto de la ampliación, pendiente de commit del usuario.

## Hallazgos que motivan el cambio

El mínimo de cinco tickets por proceso no permite demostrar I0–I6 con solo cinco incidentes. Se planifican siete incidentes y cinco tickets en cada uno de los otros seis procesos, total 37. La consulta web identificó que su resumen no contenía los umbrales originales y no afirmó haber leído los archivos. El diseño local se contrasta con el enunciado, no con supuestas capacidades del chat web.

El modelo actual de app_users solo admite admin/consulta y un rol. La ampliación necesita una migración compatible a roles múltiples, sin elevar consulta por membresía en colas. No basta con agregar opciones al frontend. El Excel CMDB no permite poblar referencias operativas completas sin revisión de los ejemplos, por lo que se distingue captura de origen y dispositivo válido.

## Documentos que orientan la implementación siguiente

- [MODELO](../MODELO.md): entidades, referencias, diagrama previo, mapeo B:U, clasificación complementaria propia y relaciones técnicas.
- [PROCESOS](../PROCESOS.md): siete rutas, permisos, cierres, reaperturas, concurrencia y reglas complementarias.
- [PRUEBAS](../PRUEBAS.md): nuevos P01–P18, cobertura adicional y aislamiento.
- [IA](../IA.md) y [prompt 008](../prompts/008-diseno-service-desk.md): evidencia real, uso limitado de la respuesta y pendientes.

## Reglas actualizadas

No tratar el diseño como implementado. No usar resultados antiguos para afirmar cobertura nueva. No completar ausencias de los ejemplos para satisfacer FK a costa de inventar información. El generador sintético usa referencias válidas y su conteo excluye software, relaciones y catálogos. Capturas deben ser reales, sin cursor ni datos de cuenta, y el recorte no modifica texto.

La clasificación complementaria, política de consulta como límite de escritura y reglas temporales son decisiones del proyecto documentadas, sujetas a revisión humana antes de implementar. Los ámbitos o categorías no enumerados en el Excel no pueden atribuirse a él.

## Validación de esta fase

Pasaron revisión de enlaces locales, git diff --check, comparación de originales y cobertura de identificadores P01–P18 como plan. No se ejecutaron pruebas funcionales nuevas. Los cambios permanecen sobre el commit base del parcial, sin commit propio ni SHA inventado.
