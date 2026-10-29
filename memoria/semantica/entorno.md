# Entorno de desarrollo (host Windows del usuario)

### El checkout Windows tiene CRLF y rompe gofmt y scripts bash
`core.autocrlf=true` y no hay `.gitattributes`: los blobs están en LF, pero la copia de trabajo tiene CRLF. Consecuencias:
- `gofmt -l .` lista **todos** los `.go` aunque nadie los haya tocado.
- `adws/trigger-github.sh` falla en Linux con `set: pipefail: invalid option name` (`adws/trigger-github.sh:2`) si se monta el checkout en un contenedor.

No es un fallo del cambio. Se distingue comprobando que el archivo no aparece en `git diff` y que `git show HEAD:<archivo> | od -c` muestra `\n`. El CI (Linux) no lo sufre.
- Estado: vigente
- Verificado: 2026-10-29
- Fuentes: [[episodica/2026-10-29-boton-color-verde]]

### gofmt -l recorre .claude/worktrees
`.claude/worktrees/` contiene otros worktrees del repositorio (ignorados vía `.git/info/exclude`). `go vet ./...` y `go test ./...` los saltan porque el directorio empieza por `.`, pero `gofmt -l .` no, así que su salida se llena de archivos ajenos. Hay que filtrar con `grep -v '^\.claude'` o lanzar `gofmt -l` sobre los directorios del proyecto.
- Estado: vigente
- Verificado: 2026-10-29
- Fuentes: [[episodica/2026-10-29-boton-color-verde]]

### El puerto 5432 puede estar ocupado por la base de datos de otro worktree
El 2026-10-29, `drag-drop-list-ordering-6c8d46-db-1`, el compose de un worktree en `.claude/worktrees/drag-drop-list-ordering-6c8d46`, publicaba el 5432. `database.yml` no permite cambiar el puerto (solo `DB_HOST`), así que la app en el host no puede usar otra base de datos en otro puerto. No hay que parar la base de datos ajena: ver [[procedimental/levantar-entorno-con-5432-ocupado]].
- Estado: vigente
- Verificado: 2026-10-29
- Fuentes: [[episodica/2026-10-29-boton-color-verde]]

### La imagen de la app compila la copia de trabajo y migra sola
El `Dockerfile` hace `COPY . .` (`Dockerfile:8`), así que `docker compose up --build` refleja los cambios sin commitear. `cmd/app/main.go:11` ejecuta `models.Migrate()` al arrancar, con lo que la base de datos de desarrollo queda migrada sin `soda`. El volumen `db-data` conserva tareas de sesiones anteriores.
- Estado: vigente
- Verificado: 2026-10-29
- Fuentes: [[episodica/2026-10-29-boton-color-verde]]

### Herramientas del host
- Disponibles: `go`, `buffalo` (`~/go/bin`), Docker 29.8.2 con Compose v5.5.1 (admite `!reset`), `node`, `gh`, `playwright-cli`.
- No disponible: Python (el alias abre la Microsoft Store). Para cálculos rápidos, usar `node -e`.
- Estado: vigente
- Verificado: 2026-10-29
- Fuentes: [[episodica/2026-10-29-boton-color-verde]]
