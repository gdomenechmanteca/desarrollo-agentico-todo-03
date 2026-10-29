# Revisión: Botón principal de color verde

- Encargo: "Cambia el color del boton a verde." (plan `plans/boton-color-verde.md`)
- Rama / commit base: `awd/boton-color-verde` / `f94784a` (`origin/main`)
- Fecha: 2026-10-29
- Veredicto: **APROBADO CON OBSERVACIONES**

## Requisitos

| ID | Requisito | Origen (usuario/plan) | Resultado | Evidencia |
|----|-----------|-----------------------|-----------|-----------|
| R1 | "Añadir" y "Guardar cambios" tienen fondo y borde verdes (`rgb(31, 122, 61)` = `#1f7a3d`) con texto blanco | usuario (verde) / plan (qué botón, tono, texto blanco) | ✅ | `evidencias/R1-inicio-reposo.png`, `evidencias/R1-editar-reposo.png`, `evidencias/R1-R2-R4-estilos-calculados.json` |
| R2 | Al pasar el ratón, fondo y borde pasan a verde oscuro (`rgb(24, 96, 47)` = `#18602f`) | plan | ✅ | `evidencias/R2-inicio-hover.png`, `evidencias/R2-editar-hover.png`, `evidencias/R1-R2-R4-estilos-calculados.json` |
| R3 | Contraste del blanco sobre el verde ≥ 4.5:1 (WCAG AA) en reposo y hover | plan | ✅ 5.37:1 y 7.63:1 | `evidencias/R3-contraste.txt` |
| R4 | No cambia nada más: punto de la marca y `+` siguen rojo teja (`rgb(183, 53, 38)`), "Cancelar" sigue gris, el anillo de foco sigue azul (`rgb(36, 90, 180)`) | plan | ✅ | `evidencias/R1-R2-R4-estilos-calculados.json`, `evidencias/R4-editar-foco.png` |
| R5 | El navegador recibe el CSS nuevo: la página enlaza `app.css?v=7` y ese recurso sirve los tokens verdes | plan | ✅ | `evidencias/R5-curl.txt` |

## Alcance

- Cambios del encargo: `public/assets/app.css:24-25` (valores de los dos tokens → R1, R2, R3) y `templates/application.plush.html:7` (`v=6` → `v=7` → R5). Ver `evidencias/alcance-git-diff.txt`.
- Cambios fuera de alcance: ninguno dentro del cambio. En la copia de trabajo hay además `.claude/skills/awd/SKILL.md` y `.claude/skills/revision/`, que **existían sin commitear antes de empezar el encargo** (creación de esta skill); no forman parte del cambio y no deben incluirse en su commit.
- Requisitos sin cubrir: ninguno.
- Exclusiones respetadas: sí. Sin cambios en `.button-primary`, `--color-primary`, `--color-focus`, `app.js?v=4`, Go, otras plantillas, modelos, migraciones ni tests.
- Tests: ninguno desactivado, borrado ni modificado.

## Estructura

Sin observaciones. El cambio reutiliza los tokens `--color-button-primary*` introducidos por `plans/boton-color-azul.md` y el patrón de cache-busting por query string del layout.

## Comandos ejecutados

| Comando | Resultado | Salida |
|---------|-----------|--------|
| `gofmt -l .` | ⚠️ lista los 17 `.go` del proyecto (ninguno tocado por el cambio) y 65 de `.claude/worktrees/` (otros worktrees, que `gofmt` recorre). Causa: CRLF en la copia de trabajo por `core.autocrlf=true` en Windows; los blobs commiteados usan LF. Preexistente y ajeno al cambio. *Corregido tras la revisión: la primera versión de este informe solo citaba `public/embed.go` y `templates/embed.go` porque se leyó únicamente el final de la salida* | `evidencias/gofmt.txt` |
| `go vet ./...` | ✅ | `evidencias/go-vet.txt` |
| `go build ./...` | ✅ | `evidencias/go-build.txt` |
| `git diff --check` | ✅ | `evidencias/alcance-git-diff.txt` |
| `soda migrate -e test` + `go test -p 1 ./...` (contenedor `golang:1.26-alpine` en la red de compose) | ✅ `actions` y `models`. ❌ `adws`: el contenedor no tenía `bash` | `evidencias/go-test.txt` |
| Mismo `go test` con `bash` instalado | ✅ `actions` y `models`. ❌ `adws/TestTriggerShellWrapper`: `set: pipefail: invalid option name` por CRLF en `adws/trigger-github.sh` de la copia Windows (el blob commiteado usa LF). Preexistente; `adws/` no se toca | `evidencias/go-test-con-bash.txt` |
| `curl` de `/` y `/assets/app.css?v=7` | ✅ | `evidencias/R5-curl.txt` |
| Playwright (`playwright-cli run-code`): estilos calculados y capturas en `/` y `/tasks/<id>/edit` | ✅ | `evidencias/R1-*.png`, `evidencias/R2-*.png`, `evidencias/R4-editar-foco.png`, `evidencias/R1-R2-R4-estilos-calculados.json` |

## Riesgos y pendientes

- **Interpretación de "el botón"**: se ha tomado como el botón principal (`.button-primary`), igual que en el encargo azul anterior. Si se refería a toda la paleta, habría que cambiar también `--color-primary`.
- **Tono de verde**: elegido sin especificación de diseño. Se ajusta cambiando solo los dos tokens.
- **Entorno**: el puerto 5432 lo ocupaba la base de datos de otro proyecto (`drag-drop-list-ordering-6c8d46-db-1`), que no se ha tocado. La app de este proyecto se arrancó en Docker con un override que no publica el puerto de la base de datos. **La app (`desarrollo-agentico-todo-03-app-1`) y su base de datos (`-db-1`) los arrancó esta revisión y siguen en marcha**; se paran con `docker compose stop`.
- Los fallos de `gofmt` y `adws` se deben al checkout CRLF de Windows. Se esperan en verde en el CI (Linux, LF), que debe confirmarlo.
