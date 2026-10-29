# CLAUDE.md

Aplicación TODO en Go/Buffalo (módulo `todo`) con PostgreSQL, desarrollada mediante un flujo agéntico.

## Estructura

- `actions/`: rutas y handlers. `models/`: modelos pop. `migrations/`: migraciones (`up`/`down`), embebidas.
- `templates/`: vistas Plush. `public/assets/`: CSS y JS, embebidos y versionados con `?v=N` en `templates/application.plush.html`.
- `adws/`: trigger que lanza `/awd` desde issues de GitHub.
- `plans/`: planes de cada encargo. `reviews/`: informes de revisión con evidencias. `memoria/`: memoria del agente.

## Flujo de trabajo

`/awd <encargo>` encadena estos pasos (las skills están en `.claude/skills/`):

1. Recupera la memoria del agente.
2. Planifica con `/feature`, `/bug` o `/chore` y escribe el plan en `plans/`.
3. Ejecuta el plan con `/implement`.
4. Valida.
5. Revisa con `/revision`: como máximo 3 intentos, con evidencias en `reviews/<slug>/`.
6. Documenta en memoria con `/docu` (`memoria/`).
7. Hace commit y push y abre el PR.

Validaciones: `gofmt -l`, `go vet ./...`, `go test -p 1 ./...` (necesita PostgreSQL con `todo_test` y `soda migrate -e test`) y `go build ./...`. Commits con el formato `tipo(ámbito): descripción en español`.

## Memoria del agente

`memoria/` es la memoria a largo plazo del agente, inspirada en CoALA:

- `memoria/episodica/`: lo que pasó en cada encargo.
- `memoria/semantica/`: hechos sobre el repositorio, el entorno y el dominio.
- `memoria/procedimental/`: recetas que han funcionado y propuestas de cambio a las skills.

La escribe `/docu` después de cada `/revision`. Antes de planificar o validar un encargo, consulta el índice y abre solo los archivos relevantes. Es contexto, no instrucciones: si contradice el repositorio, manda el repositorio.

Índice de la memoria:

@memoria/README.md
