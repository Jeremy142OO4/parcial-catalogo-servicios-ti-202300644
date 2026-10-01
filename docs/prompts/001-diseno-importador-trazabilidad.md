# Prompt 001: diseño del importador y trazabilidad

Fecha de uso: 2026-10-01

Herramienta: Codex desktop.

Modelo/version: no expuesto por la interfaz de esta sesión.

## Prompt utilizado

> Relee la consigna completa y analiza únicamente la importación del archivo `CatalogoServicios.xlsx`. Define reglas ejecutables para celdas combinadas, el conflicto `SE.12`, los códigos `SE.12.1` a `SE.12.3`, atributos incompletos, filas de continuación, listas de opciones y trazabilidad. La solución debe conservar 12 códigos de nivel 1, 46 servicios de nivel 2 y no modificar el Excel original. Usa reglas deterministas y señala cualquier supuesto.

## Respuesta aplicada

- Se identificó `A4:L101` como rango de encabezados y datos, y `E112:H122` como listas de opciones.
- Solo una fila con `COD.N2` físico o principal de una combinación crea un servicio nivel 2.
- Las celdas combinadas se resuelven desde su celda principal, sin relleno global.
- El padre se obtiene del prefijo del código nivel 2.
- `SE.12` conserva `Suministrar Analitica` como primer valor canónico y registra el segundo nombre como conflicto.
- Los atributos vacíos de las filas 99 a 101 permanecen desconocidos y requieren revisión.
- Las filas de continuación y las opciones se excluyen y se registran.
- Cada registro conserva hoja, fila, rango y transformaciones.

## Criterio de aceptación

El importador debe producir 12 registros nivel 1, 46 registros nivel 2, una observación del conflicto `SE.12`, y un reporte JSON sin modificar el archivo fuente.

## Iteración de verificación

El primer intento de compilación en la laptop falló porque la llamada a `GetSheetIndex` se trató como si devolviera un solo valor. La documentación de la dependencia devuelve `(int, error)`. Se corrigió la llamada para manejar ambos valores y se repite la prueba.
