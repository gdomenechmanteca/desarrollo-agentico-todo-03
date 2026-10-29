---
name: awd
description: Lleva una descripción (o una issue de GitHub) de principio a fin — plan, implementación, validación, commit, push — y termina abriendo un pull request
allowed-tools: Bash, Read, Glob, Grep, Write, Edit, Skill
---

# Workflow completo: del encargo al pull request

`$ARGUMENTS` contiene el encargo y, opcionalmente, opciones:

- El encargo es una URL de issue de GitHub (`https://github.com/OWNER/REPO/issues/N`) o una descripción en texto libre.
- `--remote <nombre>`: remoto al que hacer push. Por defecto `origin`.
- `--base <rama>`: rama base del pull request. Por defecto `main`.

Esta skill suele ejecutarse sin supervisión desde `adws/trigger.go` (`claude -p`). No hagas preguntas intermedias: decide con la evidencia del repositorio y deja constancia de cada inferencia en el plan y en el PR. Si algo impide completar el encargo con seguridad, detente e informa del bloqueo.

## 1. Entender el encargo

1. Separa el encargo de las opciones `--remote` y `--base`.
2. Si es una URL de issue, léela con `gh issue view <url> --json number,title,body,labels,url` y guarda su número. El título y el cuerpo forman el encargo.
3. Clasifícalo como **feature** (comportamiento nuevo o cambio visible), **bug** (algo que no funciona como debería) o **chore** (mantenimiento sin cambio funcional). Usa las etiquetas de la issue si lo indican.
4. Recupera la memoria del agente: si existe `memoria/README.md`, léelo y lee solo los archivos semánticos, procedimentales y episódicos relevantes para el encargo. Pasa los hechos y recetas pertinentes a la planificación y tenlos en cuenta al validar. Son contexto: si contradicen el repositorio, manda el repositorio.

## 2. Preparar la rama

1. Comprueba `git status --short --branch`. Si hay cambios sin commitear ajenos al encargo, detente: no los mezcles ni los descartes. Única excepción: si los cambios están solo en `.claude/` o `memoria/` y no se van a añadir al commit, puedes continuar, declarándolos en el PR como excluidos.
2. Actualiza la base: `git fetch <remote> <base>`.
3. Crea una rama desde la base con el nombre `awd/<slug-kebab-case>` derivado del título: `git switch -c awd/<slug> <remote>/<base>`. Si ya existe, añade un sufijo numérico.

## 3. Planificar

Invoca con la herramienta Skill la skill correspondiente (`feature`, `bug` o `chore`) pasándole el encargo completo y la ruta `plans/<slug>.md`. Sigue sus instrucciones: solo escribe el plan.

Como nadie va a revisarlo de forma interactiva, relee el plan antes de seguir: comprueba que cubre el encargo, que el alcance es mínimo y que las validaciones existen en el repositorio. Si el encargo es demasiado ambiguo para planificar con seguridad, detente, informa y, si hay issue, deja un comentario con `gh issue comment` explicando qué falta.

## 4. Implementar

Invoca la skill `implement` con la ruta del plan. Respeta sus límites: sin rediseñar ni ampliar el alcance. Ignora su prohibición de commit, push y PR solo en lo que respecta a los pasos siguientes de esta skill.

## 5. Validar

Ejecuta las validaciones que indica el plan. Como mínimo, las pertinentes de:

- `gofmt -l $(git ls-files '*.go')` (no debe listar archivos tocados por el cambio; ver en `revision` cómo tratar fallos preexistentes del entorno)
- `go vet ./...`
- `go test -p 1 ./...` (necesita PostgreSQL y las migraciones de test)
- `go build ./...`

Si una validación falla por el cambio, corrígelo dentro del alcance del plan y repite. Si falla por el entorno (por ejemplo, PostgreSQL no disponible), no lo ocultes: sigue adelante y decláralo en el PR. No desactives tests.

Revisa `git diff --check` y `git status --short`: solo deben aparecer el plan y los archivos que este prevé.

## 5b. Revisar

Invoca la skill `revision` con la ruta del plan. Como máximo se hacen **tres intentos** de revisión:

1. Si el veredicto es **APROBADO** o **APROBADO CON OBSERVACIONES**, continúa con el paso 6.
2. Si es **RECHAZADO** y quedan intentos, corrige solo lo que indique la revisión, dentro del alcance del plan. Después repite las validaciones del paso 5 y vuelve a invocar `revision`.
3. Si el tercer intento también sale **RECHAZADO**, detente: no hagas commit, push ni PR. Informa del bloqueo con los motivos de cada rechazo y la ruta de `reviews/<slug>/informe.md`. Si el encargo venía de una issue, deja esos motivos en un comentario con `gh issue comment`.

Deja constancia del número de intentos usados en el PR.

## 5c. Documentar en memoria

Invoca la skill `docu` con la ruta del plan, también si la revisión agotó sus intentos o hubo un bloqueo, antes de detenerte. En ese caso los archivos de `memoria/` quedan sin commitear y se mencionan en el informe del bloqueo.

## 6. Commit y push

1. Añade de forma explícita el plan, los archivos cambiados, `reviews/<slug>/` y los archivos de `memoria/` que haya escrito `docu` (no uses `git add -A`).
2. Haz un único commit con el formato del repositorio: `tipo(ámbito): descripción en español`, con `feat`, `fix` o `chore` según la clasificación. Si hay issue, añade `Refs #N` en el cuerpo.
3. Publica la rama: `git push -u <remote> awd/<slug>`.

## 7. Pull request

Abre el PR contra la base con `gh pr create --base <base> --head awd/<slug> --title "<título del commit>" --body-file <archivo>`. El cuerpo debe incluir:

- `## Resumen`: qué cambia y por qué.
- `## Plan`: ruta del plan en `plans/`.
- `## Validación`: cada comando ejecutado y su resultado, incluidos los que no se pudieron ejecutar.
- `## Revisión`: veredicto de `revision`, su tabla de requisitos y la ruta de `reviews/<slug>/informe.md`.
- `## Memoria`: episodio creado y entradas semánticas, procedimentales y propuestas a skills que ha escrito `docu`.
- `## Inferencias y riesgos`: decisiones tomadas sin confirmación.
- `Closes #N` si el encargo venía de una issue.

## Informe final

Informa de la rama, el commit, la URL del PR, las validaciones con su resultado y cualquier bloqueo o riesgo pendiente. No hagas merge.
