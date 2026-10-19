# Feature: reordenar tareas con drag and drop

## Encargo y criterios de aceptación

Encargo literal: «quiero añadir drag and drop a la lista para poder ordenarlas».

Requisitos explícitos:

- El usuario puede arrastrar una tarea de la lista y soltarla en otra posición.
- El nuevo orden es el que muestra la lista.

Inferencias (necesarias para que la funcionalidad tenga sentido o para no degradar la app):

- El orden se **persiste** en base de datos: al recargar la página, o en otro navegador, se ve el mismo orden.
- Las tareas nuevas se añaden **al final** de la lista (comportamiento actual, que ordena por `created_at`).
- Completar, reabrir o editar una tarea **no** cambia su posición.
- Funciona con ratón y con pantalla táctil (la app tiene diseño móvil explícito en `app.css`).
- Existe una alternativa por teclado (flechas sobre el asa de arrastre) para no dejar la funcionalidad fuera del alcance de quien no usa puntero.

Criterios de aceptación:

1. Cada tarea muestra un asa de arrastre con `aria-label="Mover tarea: <título>"`.
2. Arrastrar una tarea por su asa y soltarla en otra posición la recoloca inmediatamente en la lista, sin recargar.
3. Tras soltar, se envía el orden completo al servidor; al recargar `/` la lista aparece en el nuevo orden.
4. Con el foco en el asa, `ArrowUp`/`ArrowDown` mueve la tarea una posición y persiste el orden igual que el arrastre; el foco permanece en el asa.
5. Una tarea creada después de reordenar aparece al final.
6. Completar/reabrir/editar/eliminar una tarea no altera el orden relativo de las demás.
7. Si el servidor rechaza el orden (lista desactualizada, error de red), la página se recarga y muestra el orden real persistido.
8. Los datos existentes conservan su orden actual (`created_at asc, id asc`) tras la migración.

## Comportamiento actual y propuesta

Actual:

- `tasks` no tiene columna de orden (`migrations/20260901000000_create_tasks.up.fizz`).
- `TasksIndex` y la rama de error de `TasksCreate` ordenan con `tx.Order("created_at asc, id asc")` (`actions/tasks.go`).
- `templates/tasks/index.plush.html` pinta `<ul class="task-list">` con un `<li class="task">` por tarea en una rejilla de 3 columnas (toggle, título, acciones).
- `public/assets/app.js` solo gestiona `data-confirm` y el toggle optimista. No hay librerías JS ni bundler: un único archivo servido desde `public/assets` con versión en query string (`app.js?v=2`, `app.css?v=3`).
- CSRF: middleware `csrf.New` activo (desactivado en `GO_ENV=test`, por eso los tests existentes no envían token). Los formularios usan `authenticity_token`.

Propuesta (cambio vertical mínimo):

- Nueva columna `position integer NOT NULL DEFAULT 0`, rellenada en la migración según el orden actual.
- La lista se ordena por `position asc, created_at asc, id asc` (el desempate mantiene estable el orden de filas con la misma posición, p. ej. fixtures sin `position`).
- `TasksCreate` asigna `MAX(position) + 1`.
- Nuevo endpoint `POST /tasks/reorder` que recibe JSON `{"ids": ["<uuid>", …]}` con **todas** las tareas en el nuevo orden y reescribe `position = índice + 1`. Responde `204` si es válido, `422` si el conjunto de ids no coincide exactamente con las tareas existentes (falta, sobra o se repite alguna), `400` si el cuerpo no se puede interpretar.
- En la UI: asa de arrastre por tarea y JS con Pointer Events (ratón y táctil) más teclado; tras soltar, `fetch` al endpoint con el token CSRF en la cabecera `X-CSRF-Token`.

## Alcance

Incluido:

- Migración SQL con columna `position` y backfill.
- Campo `Position` en `models.Task`, cálculo de la siguiente posición y reordenación validada en el paquete `models`.
- Orden de la lista por `position`.
- Ruta y acción `POST /tasks/reorder`.
- Asa de arrastre en la plantilla, JS de arrastre/teclado/persistencia y CSS asociado.
- Tests de modelo y de acción.

Excluido:

- Librerías externas de drag and drop (SortableJS, etc.) o un bundler.
- Animaciones de reordenación más allá de un estilo visual para la tarea arrastrada.
- Separar pendientes y completadas o mover automáticamente las completadas al final.
- Reordenación parcial o por posición relativa («mover X después de Y»); siempre se envía la lista completa.
- Corregir el toggle optimista existente en `app.js`, que hace `fetch(form.action)` como **GET** sin `_method` ni token (no hay ruta GET para `/tasks/{id}/toggle`). Es un problema previo y ajeno; ver Riesgos. El código nuevo no debe copiar ese patrón.

## Diseño por capas

### Persistencia y migraciones

Nuevo par de migraciones SQL (el repo ya usa `.sql` para el seed y `.fizz` para el esquema; se elige SQL porque el backfill con `row_number()` no es expresable en fizz):

`migrations/20261019000000_add_position_to_tasks.up.sql`

```sql
ALTER TABLE tasks ADD COLUMN position integer NOT NULL DEFAULT 0;

UPDATE tasks
SET position = ordered.rn
FROM (
  SELECT id, row_number() OVER (ORDER BY created_at ASC, id ASC) AS rn
  FROM tasks
) AS ordered
WHERE tasks.id = ordered.id;
```

`migrations/20261019000000_add_position_to_tasks.down.sql`

```sql
ALTER TABLE tasks DROP COLUMN position;
```

`migrations/embed.go` ya embebe el directorio; comprobar que el patrón incluye `*.sql` (el seed lo demuestra) y no tocarlo si es así.

### Modelo (`models/task.go`)

- Añadir `Position int \`json:"position" db:"position" form:"-"\`` a `Task` (`form:"-"` para que `c.Bind` en `TasksCreate` no permita fijarla desde el formulario).
- Añadir constante `TaskOrder = "position asc, created_at asc, id asc"` para no duplicar la cadena en las acciones.
- `NextTaskPosition(tx *pop.Connection) (int, error)`: `SELECT COALESCE(MAX(position), 0) + 1 FROM tasks` vía `tx.RawQuery(...).First(&n)` (o estructura auxiliar con `db:"next"` si `First` sobre `int` no funciona con pop; usar lo que funcione y dejarlo cubierto por test).
- `var ErrInvalidTaskOrder = errors.New("el orden debe incluir cada tarea exactamente una vez")`.
- `ReorderTasks(tx *pop.Connection, ids []uuid.UUID) error`:
  1. Cargar todos los ids existentes (`tx.Select("id").All(&tasks)`).
  2. Si `len(ids) != len(existentes)`, si hay un id repetido o un id desconocido → devolver `ErrInvalidTaskOrder`.
  3. Para cada `i, id`: `tx.RawQuery("UPDATE tasks SET position = ? WHERE id = ?", i+1, id).Exec()`. No se modifica `updated_at`: reordenar no edita el contenido de la tarea.
  - Con cero tareas y `ids` vacío es válido y no hace nada.

### Acción y ruta (`actions/tasks.go`, `actions/app.go`)

- `TasksIndex` y la rama de error de `TasksCreate`: sustituir `"created_at asc, id asc"` por `models.TaskOrder`.
- `TasksCreate`: antes de `ValidateAndCreate`, `task.Position, err = models.NextTaskPosition(tx)`; propagar el error.
- Nueva acción:

```go
// TasksReorder guarda el orden de la lista recibido como lista completa de ids.
func TasksReorder(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	input := struct {
		IDs []uuid.UUID `json:"ids"`
	}{}
	if err := c.Bind(&input); err != nil {
		return c.Error(http.StatusBadRequest, err)
	}

	if err := models.ReorderTasks(tx, input.IDs); err != nil {
		if errors.Is(err, models.ErrInvalidTaskOrder) {
			return c.Error(http.StatusUnprocessableEntity, err)
		}
		return err
	}

	return c.Render(http.StatusNoContent, nil)
}
```

  Si `c.Render(204, nil)` no es válido en esta versión de Buffalo, usar `c.Response().WriteHeader(http.StatusNoContent); return nil` (detalle mecánico).
- Ruta: `app.POST("/tasks/reorder", TasksReorder)` junto a `app.POST("/tasks", TasksCreate)`. Se usa `POST` y no `PUT` para no competir con `PUT /tasks/{task_id}`.
- El middleware `popmw.Transaction` ya envuelve la petición: un error o un estado ≥ 400 revierte las actualizaciones parciales.

### Renderizado (`templates/tasks/index.plush.html`)

- En `<ul class="task-list">` añadir `data-reorder-url="/tasks/reorder"` y `data-csrf-token="<%= authenticity_token %>"`.
- En cada `<li class="task …">` añadir `data-task-id="<%= t.ID %>"`.
- Como primer hijo del `<li>`, antes del formulario de toggle:

```html
<button type="button" class="drag-handle" data-drag-handle aria-label="Mover tarea: <%= t.Title %>" title="Arrastrar para reordenar">
  <svg aria-hidden="true" viewBox="0 0 24 24" fill="currentColor"><circle cx="9" cy="6" r="1.5"/><circle cx="15" cy="6" r="1.5"/><circle cx="9" cy="12" r="1.5"/><circle cx="15" cy="12" r="1.5"/><circle cx="9" cy="18" r="1.5"/><circle cx="15" cy="18" r="1.5"/></svg>
</button>
```

- Añadir, tras el `<ul>`, una región `<p class="sr-only" data-reorder-status aria-live="polite"></p>` para anunciar movimientos por teclado.
- `templates/application.plush.html`: subir versiones a `app.css?v=4` y `app.js?v=3`.

### JavaScript (`public/assets/app.js`)

Añadir un bloque independiente (mismo estilo ES5 que el existente: `var`, `function`, sin módulos):

- **Arrastre con Pointer Events** sobre `[data-drag-handle]`:
  - `pointerdown` (solo botón principal): guardar el `li`, `setPointerCapture`, añadir clase `task-dragging` al `li` y `is-reordering` a la lista.
  - `pointermove`: recorrer los hermanos `li.task` (excepto el arrastrado); el primero cuyo punto medio vertical (`getBoundingClientRect`) esté por debajo de `clientY` es el destino → `list.insertBefore(dragged, destino)`; si ninguno, `appendChild`. Solo mover si la posición cambia.
  - `pointerup`/`pointercancel`: quitar clases; si el orden cambió respecto al inicial, llamar a `saveOrder(list)`.
- **Teclado** (`keydown` sobre `[data-drag-handle]`): `ArrowUp` intercambia con `previousElementSibling`, `ArrowDown` con `nextElementSibling`; `preventDefault`, mantener foco en el asa (`handle.focus()` tras mover), actualizar `[data-reorder-status]` con «<título> movida a la posición N de M» y llamar a `saveOrder`.
- **`saveOrder(list)`**: construir `ids` con `li.dataset.taskId` en orden DOM y hacer

```js
fetch(list.dataset.reorderUrl, {
  method: "POST",
  headers: { "Content-Type": "application/json", "X-CSRF-Token": list.dataset.csrfToken },
  body: JSON.stringify({ ids: ids }),
  credentials: "same-origin"
}).then(function (res) { if (!res.ok) { window.location.reload(); } })
  .catch(function () { window.location.reload(); });
```

- No usar la API HTML5 `draggable`/`dragstart`: no funciona de forma fiable en táctil.

### CSS (`public/assets/app.css`)

- `.task`: `grid-template-columns: 28px 44px minmax(0, 1fr) auto;` y en `@media (max-width: 760px)` `24px 36px minmax(0, 1fr) auto;`.
- `.drag-handle`: mismo reset que `.icon-action` (fondo transparente, sin borde, `color: var(--color-muted)`), `min-height: 44px`, `cursor: grab`, `touch-action: none` (imprescindible para que el arrastre táctil no haga scroll), icono de 16px; `:focus-visible` con el contorno que use el resto de botones.
- En `@media (hover: hover) and (pointer: fine)` mostrar el asa con `opacity: 0` y hacerla visible en `.task:hover`/`.task:focus-within`, igual que `.task-actions`.
- `.task-dragging`: fondo suave y sombra ligera, `cursor: grabbing`; `.is-reordering` con `user-select: none`.
- El bloque `prefers-reduced-motion` existente ya anula transiciones.

## Archivos relevantes

| Ruta | Responsabilidad |
| --- | --- |
| `migrations/20261019000000_add_position_to_tasks.up.sql` (nuevo) | Añadir `position` y rellenarla con el orden actual |
| `migrations/20261019000000_add_position_to_tasks.down.sql` (nuevo) | Eliminar `position` |
| `migrations/embed.go` | Embebe migraciones; solo verificar que incluye `.sql` |
| `models/task.go` | Campo `Position`, `TaskOrder`, `NextTaskPosition`, `ReorderTasks`, `ErrInvalidTaskOrder` |
| `models/task_test.go` | Tests de posición y reordenación |
| `actions/app.go` | Ruta `POST /tasks/reorder` |
| `actions/tasks.go` | Orden por posición, posición al crear, acción `TasksReorder` |
| `actions/tasks_test.go` | Tests de índice, creación y reordenación |
| `templates/tasks/index.plush.html` | Asa, `data-*` y región `aria-live` |
| `templates/application.plush.html` | Versión de assets |
| `public/assets/app.js` | Arrastre, teclado y persistencia |
| `public/assets/app.css` | Columna del asa y estados de arrastre |
| `fixtures/tasks.toml` | Sin cambios (las filas sin `position` toman `0` y desempatan por `created_at`) |

## Implementación paso a paso

1. Crear el par de migraciones SQL indicado y confirmar que `migrations/embed.go` las incluye.
2. En `models/task.go` añadir `Position`, `TaskOrder`, `ErrInvalidTaskOrder`, `NextTaskPosition` y `ReorderTasks` (imports `errors`, `uuid` ya disponible).
3. Escribir los tests de modelo (ver abajo) y ejecutarlos con la base de test migrada.
4. En `actions/tasks.go` usar `models.TaskOrder` en las dos consultas de lista, asignar `Position` en `TasksCreate` y añadir `TasksReorder`.
5. Registrar `app.POST("/tasks/reorder", TasksReorder)` en `actions/app.go`.
6. Escribir los tests de acción y ejecutarlos.
7. Actualizar `templates/tasks/index.plush.html` y la versión de assets en `templates/application.plush.html`.
8. Añadir el bloque de reordenación a `public/assets/app.js` y los estilos a `public/assets/app.css`.
9. Validación completa y prueba manual en navegador (ver abajo).

## Tests y validación

Requisitos: PostgreSQL accesible con la configuración `test` de `database.yml`, `GO_ENV=test` y migraciones aplicadas (`soda migrate -e test`, como en `.github/workflows/ci.yaml`), incluida la nueva.

Tests de modelo (`models/task_test.go`):

- `Test_NextTaskPosition_EmptyTable`: devuelve `1`.
- `Test_NextTaskPosition_AfterMax`: con tareas en posiciones 1 y 5 devuelve `6`.
- `Test_ReorderTasks_PersistsPositions`: tres tareas; reordenar `[c, a, b]` → posiciones `c=1, a=2, b=3`.
- `Test_ReorderTasks_MissingID_Invalid`, `…_UnknownID_Invalid`, `…_DuplicateID_Invalid`: devuelven `ErrInvalidTaskOrder` (`errors.Is`) y las posiciones no cambian.

Tests de acción (`actions/tasks_test.go`):

- `Test_TasksIndex_OrdersByPosition`: crear tareas con posiciones inversas a su creación; comprobar con `strings.Index` que en el HTML aparecen en orden de `position`.
- `Test_TasksIndex_RendersDragHandle`: el cuerpo contiene `data-reorder-url="/tasks/reorder"`, `data-task-id="<id>"` y `Mover tarea: <título>`.
- `Test_TasksCreate_AppendsAtEnd`: con una tarea existente en posición 7, la nueva queda en 8.
- `Test_TasksReorder_Persists`: `as.JSON("/tasks/reorder").Post(map[string]interface{}{"ids": [...]})` → `204`; recargar y comprobar posiciones; `GET /` muestra el nuevo orden.
- `Test_TasksReorder_IncompleteList`: falta un id → `422`; posiciones intactas.
- `Test_TasksReorder_UnknownID`: id aleatorio → `422`.
- `Test_TasksReorder_InvalidUUID`: `{"ids": ["no-es-un-uuid"]}` → `400`.
- Los tests existentes deben seguir pasando sin cambios (en especial los que usan el fixture `tres tareas`).

Comandos:

```bash
gofmt -l .
go vet ./...
go test -p 1 ./...
go build ./...
```

Prueba manual (con `/env:start`, en escritorio y con viewport móvil): arrastrar una tarea, recargar y comprobar el orden; mover con flechas desde el asa; crear una tarea y verificar que va al final; completar una tarea y verificar que no se mueve; comprobar en la consola de red que `POST /tasks/reorder` responde `204` (en desarrollo el CSRF está activo, así que esto valida el envío del token).

## Riesgos y recuperación

- **Token CSRF**: en tests el middleware está desactivado, así que solo la prueba manual demuestra que `X-CSRF-Token` es aceptado. Si el endpoint devuelve 403 en desarrollo, revisar que el valor de `data-csrf-token` es el `authenticity_token` de la misma petición.
- **Orden concurrente** (dos pestañas): el último envío gana; si una pestaña tiene una lista desactualizada (tarea creada o borrada en otra), el servidor responde `422` y el JS recarga. Aceptado.
- **Toggle existente**: el `fetch(form.action)` del toggle es un GET sin ruta, por lo que el cambio optimista de completado probablemente no se persiste hoy. No se corrige aquí; si se confirma, abrir un `/bug` aparte. No interfiere con el orden.
- **Pop y `First` sobre `int`** en `NextTaskPosition`: si falla, usar una estructura auxiliar; el test lo detecta.
- **Rendimiento**: `ReorderTasks` hace un `UPDATE` por tarea. Suficiente para una lista personal; si creciera, sustituir por un único `UPDATE … FROM (VALUES …)`.
- **Recuperación**: la migración es reversible (`soda migrate down -e <env>` elimina la columna). Revertir el código restaura el orden por `created_at`; la columna `position` sobrante no afecta a la versión anterior porque pop solo usa columnas declaradas en el struct.
