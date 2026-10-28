# Contexto de permisos y actores administrativos

Fecha: 2026-10-28. Cambio motivado por la segunda auditoría y la solicitud de Jeremy de verificar muy bien los permisos.

El panel antiguo no utilizaba los métodos nuevos con actor. Sus endpoints ahora llaman `CreateUserBy`, `UpdateUserBy` y `SetUserActiveBy`, usando el usuario autenticado y una transacción. La migración 010 incorpora snapshots de usuarios sin hashes de contraseñas. Los procesos internos sin actor conservan origen sistema, sin atribuciones humanas inventadas.

Validar permisos antes de esperar un bloqueo de cola no garantiza su vigencia después de la espera. La nueva regla es bloquear primero y evaluar después, manteniendo el bloqueo hasta commit. Se aplica a las operaciones de tickets que dependen de grants. El actor administrativo también se bloquea antes de consultar sus roles en una sentencia independiente, evitando reutilizar un snapshot anterior a la espera.

Se verificaron seis revocaciones concurrentes por usuario o rol, el panel antiguo mediante HTTP, la negación de consulta con roles elevados y una revocación concurrente del administrador. El control negativo en una copia temporal demuestra que el test detecta el orden defectuoso. [Evidencia real y límites](../evidencias/service-desk-permisos-y-actor.md).

No se hizo commit, push ni despliegue sobre la demo original. No convertir estos resultados en una declaración global de tarea completa ni atribuirlos al commit base sin los cambios del worktree.
