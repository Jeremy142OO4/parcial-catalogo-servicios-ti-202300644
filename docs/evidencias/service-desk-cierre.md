# Cierre técnico y versión de entrega

Fecha: 28 de octubre de 2026, America/Guatemala. Este registro amplía la [verificación P01–P18](service-desk-final.md), no reemplaza ni altera sus salidas históricas.

## Cambios de cierre

- Vista administrativa **CMDB → Origen del Excel**, con las dos filas, 20 campos por fila, hash y coordenadas. Carga repetible de origen también desde la preparación de demostración.
- Prueba de que conservar origen no crea CI ni relaciones ficticias. Comparación de los 40 valores con las celdas originales, conservadas como cadenas de lectura de Excelize, no como tipos nativos de la celda.
- Negativas explícitas de cambio sin aprobación, resolución vacía y cierre sin confirmación, con actor de cierre registrado.
- Build con hashes de originales y búsqueda heurística de claves privadas o tokens. Este control no es una auditoría completa de secretos.
- Cuatro consultas web reales y dos mejoras comparadas en [013–016](../IA.md). Nueve registros de prompts de esta ampliación en total.
- [Guía de defensa](../DEFENSA-SERVICE-DESK.md) y cuentas preparadas, sin subir diapositivas docentes ni mover `parcial-v2.0`.

## Ejecuciones previas al commit

[Runner de cierre](service-desk-cierre-salida.txt), [persistencia](service-desk-cierre-persistencia.txt) y [demostración preservada](service-desk-cierre-demo.txt) terminaron con código 0. Proyecto `catalogo-final-test`, base `catalogo_desk_test`, puertos 18180/18181. La demostración está en `catalogo-service-desk-demo`, base distinta, puerto 18090.

La laptop tenía los contenedores de demostración detenidos al comenzar esta sesión. Se iniciaron sus servicios existentes y se actualizaron API/frontend, conservando volúmenes. El recibo de demostración muestra los mismos digest de tickets e historial antes y después y mantiene 37 tickets, siete colas, ocho FAQ y 5,000 dispositivos. La nueva fuente solo agrega sus dos registros de trazabilidad.

La prueba inicial de origen comprobaba número de filas y campos. Tras la revisión web se amplió para comparar cada valor y los conteos operativos. La ejecución desde clon limpio debe verificar esa ampliación y asociarse al SHA real del código.

![Origen del Excel en la aplicación](../Imagenes/service-desk-origen-excel.png)

## Reproducibilidad del commit

Pendiente de registrar el SHA creado y la ejecución desde un clon limpio de ese SHA. No se declara cumplido por las pruebas del worktree.

## Límites que requieren intervención humana

P04 sigue parcial en la fidelidad exacta de la jerarquía docente. Las pruebas de padres y ciclos pasan, pero no se atribuyen las seis subclases complementarias a una fuente que no las enumera. Se solicitó al usuario confirmar la interpretación comunicada por el catedrático.

No es un requisito expreso tener un editor genérico de esquemas de atributos. Los atributos dependientes de clase sí se validan y editan. Los ejemplos con referencias incompletas no se convierten automáticamente en CI operativos. La entrega en Wedy, la hora límite definitiva y la defensa oral corresponden al estudiante.
