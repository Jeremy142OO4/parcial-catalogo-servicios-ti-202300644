# Identidad técnica estable y administración auditable

Fecha: 2026-10-28. Decisión propia motivada por la auditoría y autorizada por Jeremy.

Un código visible de clasificación es editable y no puede ser la identidad de una regla técnica. Se incorpora `technical_profile` inmutable a las seis clases reservadas, con unicidad y validación en PostgreSQL. Los identificadores de CI y las claves de generación no cambian al renombrar la clase. Los perfiles se asignan al migrar las clases originales, no desde el formulario.

La trazabilidad administrativa se implementa con triggers dentro de la misma transacción que la modificación, para conservar también eliminaciones. La aplicación establece el actor en una variable local de transacción, no en un dato libre enviado por el cliente. Los procesos internos sin actor quedan identificados como sistema. El historial conserva JSON anterior y posterior, fecha y entidad, sin FK a una relación eliminable. Las relaciones y los vínculos con servicios se reflejan además en el historial de los CI afectados.

Administrar una cola permite editarla y gestionar todos sus permisos, incluida la delegación a otro usuario o rol, únicamente dentro de esa cola. No implica atender, transferir ni aprobar tickets. No permite crear colas ni alterar roles globales. El administrador global mantiene esas facultades y consulta nunca escribe. Se revalida el permiso después del bloqueo de la cola, de modo que una revocación concurrente no deje autorizada una configuración con permisos anteriores.

Los cambios nuevos se prueban en un entorno remoto separado. Las migraciones de una base existente se verifican sobre una copia, sin convertir los resultados de fixtures en conteos de demostración. Ver [resultados y límites](../evidencias/service-desk-correcciones-auditoria.md).
