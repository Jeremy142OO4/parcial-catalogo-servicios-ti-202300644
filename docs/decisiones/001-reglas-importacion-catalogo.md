# Decisión 0001: reglas iniciales de importación

Fecha: 2026-10-01

## Decisiones

1. El rango de datos es `A4:L101` y el rango de opciones es `E112:H122`.
2. Una fila de nivel 2 solo se crea cuando `COD.N2` existe físicamente en la fila o es la celda principal de una combinación. Las filas hijas de una combinación no crean duplicados.
3. Los valores de celdas combinadas se recuperan desde la celda principal y solo dentro del rango declarado.
4. La relación padre de un servicio nivel 2 se obtiene del prefijo de su código. Por ejemplo, `SE.12.3` pertenece a `SE.12`. Esto permite resolver la fila 101 sin copiar valores de la fila 100.
5. Para el conflicto `SE.12`, se conserva como nombre canónico el primer valor encontrado: `Suministrar Analitica` de la fila 99. `Mantener Tableros de Control` de la fila 100 se conserva como evidencia y se reporta como observación.
6. `SE.12.1`, `SE.12.2` y `SE.12.3` se conservan como texto.
7. Los atributos vacíos de las filas 99 a 101 se guardan como desconocidos (`NULL`) y los servicios quedan marcados para revisión.
8. Las filas sin `COD.N2` se registran como filas ignoradas de continuación o no-servicio. Las opciones de `E112:H122` se excluyen del catálogo.

## Justificación

Las reglas son deterministas, repetibles y conservan la información de origen. No se realiza un relleno global hacia abajo y no se inventan valores faltantes.
