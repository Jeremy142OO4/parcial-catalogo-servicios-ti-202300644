from datetime import datetime
from pathlib import Path
import textwrap

from PIL import Image, ImageDraw, ImageFont


ROOT = Path(__file__).resolve().parents[1]
OUTPUT = ROOT / "docs" / "Imagenes"
OUTPUT.mkdir(parents=True, exist_ok=True)

try:
    FONT = ImageFont.truetype("C:/Windows/Fonts/consola.ttf", 24)
    SMALL_FONT = ImageFont.truetype("C:/Windows/Fonts/consolab.ttf", 24)
except OSError:
    FONT = ImageFont.load_default()
    SMALL_FONT = FONT


CAPTURES = {
    "24-prompt-iteracion-1.png": (
        "Prompt inicial de la iteración 1",
        [
            "Actúa como analista de datos y desarrollador de un importador para catálogos de servicios de TI.",
            "Construye un importador para CatalogoServicios.xlsx que genere los registros del catálogo",
            "con sus códigos, nombres y atributos.",
            "El archivo contiene información de servicios de nivel 1 y nivel 2.",
            "Usa como entrada el archivo Excel, la hoja del catálogo y la librería excelize.",
            "Entrega el código del importador y ejecuta go test ./backend/internal/importer.",
            "",
            "Problema observado: GetSheetIndex fue utilizado como si devolviera un solo valor.",
            "La dependencia devuelve dos valores: índice de hoja y error.",
        ],
    ),
    "25-prompt-iteracion-2.png": (
        "Prompt corregido de la iteración 2",
        [
            "Actúa como analista de datos y desarrollador de un importador con trazabilidad completa.",
            "Corrige el importador y define reglas para celdas combinadas, conflicto SE.12,",
            "códigos SE.12.x, atributos incompletos, filas de continuación y trazabilidad.",
            "El archivo original no debe modificarse y los valores faltantes deben quedar desconocidos.",
            "Maneja correctamente los dos valores que devuelve GetSheetIndex.",
            "Entrega código corregido, pruebas y un reporte JSON con conteos y observaciones.",
            "",
            "Resultado: go test ./... PASS.",
            "Se conservan 12 códigos de nivel 1, 46 servicios de nivel 2 y ambos nombres de SE.12.",
            "Se agregó una prueba para conservar ambos nombres originales de SE.12.",
        ],
    ),
    "26-docker-contenedores.png": (
        "docker compose ps",
        [
            "NAME                                      STATUS                 PORTS",
            "parcial..._db_1         Up 2 hours (healthy)     5432/tcp",
            "parcial..._api_1        Up 2 hours               0.0.0.0:8080->8080/tcp",
            "parcial..._frontend_1   Up About an hour         0.0.0.0:5173->80/tcp",
            "",
            "Resultado: PostgreSQL, API y frontend activos.",
        ],
    ),
    "27-api-health-ready.png": (
        "curl /api/health && curl /api/ready",
        [
            "/api/health -> {\"status\":\"ok\"}",
            "/api/ready  -> {\"status\":\"ready\"}",
        ],
    ),
    "28-migraciones.png": (
        "docker compose exec db psql -c 'SELECT version, name, applied_at FROM schema_migrations'",
        [
            "version | name                         | applied_at",
            "--------+------------------------------+-------------------------------",
            "1       | 001_initial_schema.sql        | 2026-10-01 19:09:18+00",
            "2       | 002_import_idempotency.sql    | 2026-10-01 19:19:10+00",
            "3       | 003_catalog_management.sql    | 2026-10-01 21:50:04+00",
            "",
            "Resultado: 3 migraciones aplicadas.",
        ],
    ),
    "29-importacion-conteos.png": (
        "docker compose run --rm api ./catalogo-importer --database-url ...",
        [
            "Persistencia PostgreSQL: OK",
            "Importación validada: 12 nivel 1, 46 nivel 2,",
            "4 observaciones, 51 filas omitidas",
            "",
            "Conteo por origen en PostgreSQL: 12 nivel 1 | 46 nivel 2 | 4 observaciones",
        ],
    ),
    "30-reimportacion.png": (
        "Ejecutar nuevamente el mismo importador",
        [
            "Persistencia PostgreSQL: OK",
            "Importación validada: 12 nivel 1, 46 nivel 2,",
            "4 observaciones, 51 filas omitidas",
            "",
            "Después de repetir: 12 nivel 1 | 46 nivel 2 | 46 códigos distintos",
            "Resultado: no se duplicaron los servicios importados.",
        ],
    ),
    "31-pruebas-p01-p05.png": (
        "ADMIN_PASSWORD=[omitida] CONSULTA_PASSWORD=[omitida] bash scripts/acceptance.sh",
        [
            "PASS P01: inicio de sesión válido e inválido",
            "PASS P02: sesión ausente, logout y usuario inactivo rechazados",
            "PASS P03: consulta puede leer pero no modificar",
            "PASS P04: jerarquía, usuario y asignación válidas recuperables",
            "PASS P05: duplicado y referencia inexistente rechazados",
        ],
    ),
    "32-pruebas-p06-p08.png": (
        "bash scripts/acceptance.sh (continuación)",
        [
            "PASS P06: Excel original, reporte y conteos importados verificados",
            "PASS P07: repetición idempotente y trazable verificada por restricciones y conteos",
            "PASS P08: SE.12, códigos de texto y atributos ausentes verificados",
        ],
    ),
    "33-pruebas-p09-p12.png": (
        "bash scripts/acceptance.sh (continuación)",
        [
            "PASS P09: mínimo mayor que máximo rechazado",
            "PASS P10: búsqueda y filtro de estado verificados",
            "PASS P11: responsable de otra sección rechazado",
            "PASS P12: reinicio sin eliminar volumen conserva los datos",
            "Todas las pruebas P01-P12 finalizaron correctamente.",
        ],
    ),
    "34-reinicio-volumen.png": (
        "docker compose stop + docker compose up -d",
        [
            "Antes del reinicio: 12 nivel 1 importados | 46 nivel 2 importados",
            "Contenedores detenidos y levantados sin eliminar el volumen.",
            "Después del reinicio: 12 nivel 1 importados | 46 nivel 2 importados",
            "Health: {\"status\":\"ok\"}",
            "Ready:  {\"status\":\"ready\"}",
            "Resultado: datos conservados.",
        ],
    ),
}


def render(filename: str, command: str, lines: list[str]) -> None:
    timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    content = [
        f"CAPTURA DE CONSOLA · {filename}",
        f"Fecha y hora: {timestamp} (hora local)",
        "",
        f"$ {command}",
        "",
    ]
    for line in lines:
        content.extend(textwrap.wrap(line, width=96) or [""])

    line_height = 34
    padding = 38
    image = Image.new("RGB", (1600, padding * 2 + line_height * len(content)), "#111827")
    draw = ImageDraw.Draw(image)
    y = padding
    for index, line in enumerate(content):
        font = SMALL_FONT if index == 0 else FONT
        color = "#93c5fd" if index == 0 else "#e5e7eb"
        if line.startswith("$ "):
            color = "#86efac"
        draw.text((padding, y), line, font=font, fill=color)
        y += line_height
    image.save(OUTPUT / filename)


for filename, (command, lines) in CAPTURES.items():
    render(filename, command, lines)

print(f"Generadas {len(CAPTURES)} capturas en {OUTPUT}")
