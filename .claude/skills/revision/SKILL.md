---
name: revision
description: Revisa tras la implementación la estructura, la funcionalidad y el ajuste estricto al encargo, y deja evidencias reproducibles de que funciona lo pedido
allowed-tools: Bash, Read, Glob, Grep, Write, Edit, Skill
---

# Revisa una implementación y deja evidencias

Se ejecuta **después** de `/implement` (o de cualquier implementación) y **antes** del commit. Su objetivo es demostrar, con evidencias guardadas en el repositorio, que se ha hecho exactamente lo pedido: ni menos ni más.

`$ARGUMENTS` identifica el encargo y, opcionalmente, el plan:

- Ruta a un plan en `plans/<slug>.md`, URL de issue de GitHub o descripción en texto libre.
- Si solo se da el plan, el encargo original es el que figura en el propio plan.
- Si no se da nada, usa el último plan modificado en `plans/` y decláralo.

El `<slug>` es el del plan; si no hay plan, derívalo en kebab-case del encargo.

No hagas preguntas intermedias: puede ejecutarse sin supervisión desde `/awd`. Revisa, no rediseñes: esta skill no corrige el código (ver el apartado *Veredicto*).

## 1. Fijar lo que se pidió

1. Lee el encargo literal (issue con `gh issue view <url> --json title,body`, texto o sección de encargo del plan) y el plan completo.
2. Extrae una lista numerada de **requisitos verificables** (`R1`, `R2`, …), cada uno redactado como comportamiento observable. Separa los que pidió el usuario de los que añadió el plan como decisión de diseño.
3. Anota las **exclusiones**: lo que el encargo o el plan dejan fuera explícitamente.

## 2. Revisar el alcance (estrictamente lo pedido)

1. Ejecuta `git status --short --branch`, `git diff --stat <base>...HEAD` y `git diff` (incluye lo no commiteado; la base por defecto es `main`).
2. Asocia cada archivo y cada bloque cambiado a un requisito. Clasifica cualquier cambio sin requisito como **fuera de alcance** (refactors, renombrados, formateo de código ajeno, dependencias nuevas, cambios de texto o estilo no pedidos, archivos de depuración).
3. Comprueba que ningún requisito ha quedado sin cambio que lo cubra y que no se ha tocado nada de las exclusiones.
4. Comprueba que no se han desactivado, borrado ni debilitado tests.

## 3. Revisar la estructura

Compara el cambio con el código vecino y con las convenciones del repositorio (Go/Buffalo):

- Ubicación: acciones en `actions/`, modelos en `models/`, vistas en `templates/`, migraciones en `migrations/` con su par `up`/`down`, assets en `public/`.
- Naming, idioma de textos y comentarios, y estilo coherentes con los archivos que se tocan.
- Migraciones reversibles y modelos con validaciones donde el patrón existente las tiene.
- Tests añadidos junto a los existentes (`*_test.go`) cubriendo cada requisito con lógica.
- Sin código muerto, `TODO` nuevos, `fmt.Println` de depuración ni secretos.

Ejecuta y registra la salida de:

- `gofmt -l .` (no debe listar archivos)
- `go vet ./...`
- `go build ./...`
- `git diff --check`

## 4. Verificar la funcionalidad

1. Ejecuta `go test -p 1 ./...` (necesita PostgreSQL y las migraciones de test). Si el entorno no lo permite, regístralo como **no verificado**, nunca como superado.
2. Arranca la aplicación siguiendo `.claude/commands/env/start.md` si no está ya respondiendo en `http://localhost:3000/healthz`. No mates procesos ajenos.
3. Para **cada requisito** diseña una prueba que lo demuestre desde fuera, como lo haría el usuario:
   - Cambios visibles en la interfaz: usa la skill `playwright-cli` para recorrer el flujo y guarda una captura antes y después de la acción clave.
   - Cambios de comportamiento HTTP: `curl -i` contra la ruta, guardando petición y respuesta.
   - Persistencia: comprueba el dato tras recargar la página o consultando la base de datos con `docker compose exec db psql`.
4. Prueba también el camino negativo o límite que el requisito implique (campo vacío, valor inválido, opción no marcada) y que la funcionalidad existente cercana sigue funcionando.
5. Deja la aplicación como la encontraste: si la arrancaste tú, indícalo en el informe para que se pueda parar.

## 5. Guardar las evidencias

Todo va a `reviews/<slug>/`:

- `informe.md`: el informe de revisión (formato abajo).
- `evidencias/`: capturas (`R1-antes.png`, `R1-despues.png`, …), salidas de comandos (`go-test.txt`, `go-vet.txt`, `gofmt.txt`, `R2-curl.txt`, …) y cualquier volcado de base de datos usado.

Cada evidencia se nombra con el requisito que demuestra. Las salidas se guardan completas, sin editar. No guardes secretos, tokens ni datos personales reales.

### Formato de `informe.md`

```markdown
# Revisión: <título del encargo>

- Encargo: <issue, plan o texto>
- Rama / commit base: <rama> / <sha>
- Fecha: <AAAA-MM-DD>
- Intento: <N>/3
- Veredicto: APROBADO | APROBADO CON OBSERVACIONES | RECHAZADO | RECHAZADO (intentos agotados)

## Requisitos

| ID | Requisito | Origen (usuario/plan) | Resultado | Evidencia |
|----|-----------|-----------------------|-----------|-----------|
| R1 | ...       | usuario               | ✅ / ❌ / ⚠️ no verificado | evidencias/R1-despues.png |

## Alcance

- Cambios fuera de alcance: <lista con archivo:línea, o "ninguno">
- Requisitos sin cubrir: <lista o "ninguno">
- Exclusiones respetadas: sí / no (detalle)

## Estructura

<hallazgos con archivo:línea, o "sin observaciones">

## Comandos ejecutados

| Comando | Resultado | Salida |
|---------|-----------|--------|
| go test -p 1 ./... | ✅ / ❌ / no ejecutado (motivo) | evidencias/go-test.txt |

## Riesgos y pendientes

<lo que no se pudo verificar y por qué>

## Historial de intentos

| Intento | Veredicto | Motivos del rechazo |
|---------|-----------|---------------------|
| 1/3     | RECHAZADO | ...                 |
```

## Veredicto

- **APROBADO**: todos los requisitos demostrados con evidencia, sin cambios fuera de alcance y con todas las comprobaciones en verde.
- **APROBADO CON OBSERVACIONES**: requisitos demostrados, pero hay hallazgos menores de estructura o algo no verificable por el entorno (declarado).
- **RECHAZADO**: algún requisito falla o no está cubierto, hay cambios fuera de alcance, o falla un test, `go vet`, `gofmt` o la compilación.

Esta skill no corrige el código. Si el veredicto es RECHAZADO, enumera exactamente qué hay que cambiar para que quien la invocó lo corrija dentro del alcance del plan y vuelva a ejecutar `/revision`.

### Límite de intentos

Para el mismo `<slug>` se admiten como máximo **tres revisiones**. El número de intento lo lleva `reviews/<slug>/informe.md`:

- Antes de empezar, si existe un informe anterior, lee su línea `- Intento: N/3`. Este será el intento `N+1`; si no hay informe, es el `1`.
- Si el informe anterior está aprobado (con o sin observaciones), es una revisión nueva del mismo encargo y vuelve a empezar en `1`.
- Cada ejecución sobrescribe `reviews/<slug>/`, pero conserva en el informe una sección `## Historial de intentos` con el veredicto y los motivos de cada rechazo anterior.
- Si el intento 3 sale **RECHAZADO**, el veredicto es definitivo: indícalo como `RECHAZADO (intentos agotados)` y no propongas otra revisión. A partir de ahí decide una persona.
- No ejecutes nunca un cuarto intento: si el informe anterior ya marca `3/3` con rechazo, detente e informa sin revisar.

No hagas commit, push ni merge. Los archivos de `reviews/<slug>/` quedan listos para incluirse en el commit del cambio.

## Informe final

Responde con el veredicto, la tabla de requisitos, los cambios fuera de alcance detectados, los comandos con su resultado y la ruta de `reviews/<slug>/informe.md`.
