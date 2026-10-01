# Frontend

Interfaz React + Vite. Incluye login local, búsqueda de servicios, consulta de la jerarquía, mantenimiento administrativo de unidades y usuarios, y asignación de responsables.

En desarrollo puede ejecutarse con `npm install` y `npm run dev`; el proxy de Vite envía `/api` a `http://localhost:8080`. En Docker Compose se sirve mediante Nginx en el puerto `5173` y se comunica con la API a través del proxy `/api`.
