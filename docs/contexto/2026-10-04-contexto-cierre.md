# Contexto del cierre de auditoría

Fecha: 2026-10-04. Integrante: Jeremy Estuardo Orellana Aldana, 202300644.

La auditoría mostró que compilar y pasar P01–P12 no garantizaba cubrir todo el alcance. Se añadió mantenimiento de catálogos auxiliares, protección de etiquetas originales, rechazo de referencias desconocidas, estado de nivel 1 y paginación. El diccionario ahora cubre claves, tipos y columnas del Excel.

Las pruebas deben leer el reporte generado en su ejecución, usar un usuario aislado para comprobar inactividad y verificar conteos de origen, métricas ausentes y persistencia del registro recién creado. La importación reporta cero creados y 58 actualizados al repetirse. El seed carga el catálogo antes de crear tres asignaciones.

El runner usa Python, curl y Bash dentro del contenedor. Las diferencias de etiquetas entre Podman Compose y Docker Compose impidieron reutilizar algunas operaciones Compose dentro del runner. Se localizaron los IDs de los contenedores de este proyecto y se usan operaciones del motor para ejecutar y reiniciar, sin eliminar volúmenes.

Las imágenes 24 a 34 son reconstrucciones gráficas y no capturas directas. Deben identificarse como tales. La evidencia nueva registra salida real de los controles y no se atribuye al modelo del ciclo anterior ni a una fecha anterior.

El commit previo es `751dbd8`, con tag sincronizado al comenzar el cierre. El integrante debe publicar las modificaciones finales y mover el tag después de su commit. El acceso del catedrático requiere confirmación en GitHub.
