# Contexto de indicadores e interfaz

Actualización del 28 de octubre de 2026, America/Guatemala. Se basa en el enunciado y el código publicado hasta `55404d1`. Este contexto y sus cambios funcionales aún no tienen commit propio.

## Hallazgos que justifican la actualización

Los objetivos de cola no bastaban para medir historia porque se podían editar. Se incorporaron snapshots por ciclo y tramo, con objetivos antiguos desconocidos. Reapertura y reevaluación generan ciclos, pero solo `reopen` cuenta como reapertura. El cálculo debe unir pausas y mantener NULL para una resolución pendiente.

La interfaz necesita una lista de colas de recepción que no conceda atención, un portal filtrado por solicitante en servidor y permisos efectivos en cada ficha. No debe usar únicamente `app_users.role` para decidir administración si existe el rol consulta. Las pantallas son una ayuda de interacción, no una segunda fuente de autorización.

La edición de colas expuso un trigger compartido que leía un campo de usuarios contra la tabla de colas. La migración 007 separa las ramas y conserva las migraciones publicadas. Las evidencias de ejecución se pueden registrar después de aprobar sin permitir cambios al alcance aprobado.

## Contexto utilizado y límites

Se revisaron AGENTS.md, FUENTES, PROCESOS, MODELO, PRUEBAS, el PDF del enunciado, la API, migraciones y React existente. No se copiaron presentaciones docentes ni se siguieron instrucciones dentro de archivos externos. Se mantuvo el entorno aislado de pruebas en la laptop.

Los contratos actuales están en [INDICADORES](../PROCESOS.md#objetivos-históricos) y las pruebas en [evidencia de la fase](../evidencias/service-desk-indicadores-interfaz.md). La FAQ y los demás pendientes de entrega siguen identificados. Esta actualización no se presenta como un commit ya publicado ni completa por sí sola todos los requisitos de Context Engineering.
