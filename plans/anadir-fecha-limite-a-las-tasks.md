# Feature: fecha límite en las tareas

## Encargo y criterios de aceptación

Encargo literal (issue #5, https://github.com/gdomenechmanteca/desarrollo-agentico-todo-03/issues/5): «quiero añadir fecha limite a las tasks». Detalle: «Vale con que sea solo un dia sin hora».

Requisitos explícitos:

- Una tarea puede tener una fecha límite.
- La fecha límite es solo un día (sin hora).

Inferencias (tomadas sin confirmación; el workflow es desatendido):

- La fecha límite es **opcional**: las tareas existentes y las nuevas sin fecha siguen siendo válidas.
- Se puede fijar al **crear** la tarea (formulario de la lista) y fijar, cambiar o **quitar** al **editarla**.
- Se muestra en la lista, debajo del título, en formato español `dd/mm/aaaa`.
- Se permiten fechas pasadas (si no, una tarea vencida no podría editarse sin cambiar su fecha).
- Una fecha mal formada se rechaza con un error de validación, igual que el título.
- No se cambia el orden de la lista (sigue siendo `models.TaskOrder`).

Criterios de aceptación:

1. El formulario de creación tiene un campo `<input type="date" name="due_date">` con etiqueta accesible «Fecha límite (opcional)».
2. Crear una tarea con `due_date=2026-11-03` la persiste con esa fecha; sin `due_date` (o vacía) se persiste sin fecha (`NULL`).
3. Crear con `due_date` inválida (p. ej. `2026-13-45`) responde `422`, muestra «La fecha límite no es válida.» y no crea la tarea.
4. La lista muestra, para cada tarea con fecha, `Fecha límite: <time datetime="2026-11-03">03/11/2026</time>`; las tareas sin fecha no muestran nada nuevo.
5. El formulario de edición precarga la fecha (`value="2026-11-03"`); guardar con otra fecha la actualiza; guardar con el campo vacío la elimina.
6. Editar con fecha inválida responde `422` con el mensaje anterior y no modifica la tarea.
7. Editar el título o la fecha no cambia `completed`, `position`, `id` ni `created_at`.
8. Las tareas existentes quedan sin fecha tras la migración.

## Comportamiento actual y propuesta

Actual:

- Tabla `tasks`: `id`, `title`, `completed`, `position`, timestamps (`migrations/20260901000000_create_tasks.up.fizz`, `migrations/20261019000000_add_position_to_tasks.up.sql`). `migrations/embed.go` embebe `*.fizz *.sql`.
- `models.Task` (`models/task.go`) valida solo el título.
- `TasksCreate` (`actions/tasks.go`) hace `c.Bind(task)` sobre `models.Task`; `TasksUpdate` enlaza un struct anónimo con solo `title`.
- `templates/tasks/index.plush.html` y `templates/tasks/edit.plush.html` muestran el error en `<p id="title-error">` iterando `errors.Get("title")`.
- CSS en `public/assets/app.css`, enlazado como `app.css?v=5` en `templates/application.plush.html`.

Propuesta (cambio vertical mínimo):

- Columna nueva `due_date date NULL`.
- Campo `DueDate *time.Time` en el modelo (nil = sin fecha), con helpers para parsear `YYYY-MM-DD` y formatear para el input y para la lista.
- Las acciones de crear y actualizar leen `due_date` como cadena, la parsean y, si es inválida, devuelven `422` con el error en la clave `due_date`.
- Plantillas: input de fecha en ambos formularios, fecha visible en la lista, errores de `due_date` mostrados junto a los de título.

## Alcance

Incluido:

- Migración SQL (up/down) de la columna.
- Campo y helpers en `models/task.go`.
- Lectura/validación de `due_date` en `TasksCreate` y `TasksUpdate`.
- Input de fecha en creación y edición; visualización en la lista; CSS mínimo y bump de versión del CSS.
- Tests de modelo y de acción.

Excluido:

- Resaltar tareas vencidas o próximas a vencer, ordenar o filtrar por fecha (requiere decidir zona horaria del «hoy»; posible issue posterior).
- Hora en la fecha límite.
- Cambios en `public/assets/app.js`, en el reordenado o en el toggle.
- Nuevas dependencias (no se usa `gobuffalo/nulls`).

## Diseño por capas

### Persistencia y migraciones

`migrations/20261026000000_add_due_date_to_tasks.up.sql`

```sql
ALTER TABLE tasks ADD COLUMN due_date date;
```

`migrations/20261026000000_add_due_date_to_tasks.down.sql`

```sql
ALTER TABLE tasks DROP COLUMN due_date;
```

Tipo `date` de PostgreSQL: solo día, sin hora ni zona. Nullable, sin default. `fixtures/tasks.toml` no se toca (las filas quedan con `NULL`).

### Modelo (`models/task.go`)

- Campo en `Task`, después de `Position`:

  ```go
  DueDate   *time.Time `json:"due_date" db:"due_date" form:"-"`
  ```

  `form:"-"` evita que `c.Bind` lo rellene; lo gestionan las acciones.

- Constantes y error:

  ```go
  // DueDateLayout es el formato de la fecha límite en formularios (input type="date").
  const DueDateLayout = "2006-01-02"

  // ErrInvalidDueDate indica que la fecha límite no tiene el formato AAAA-MM-DD.
  var ErrInvalidDueDate = errors.New("La fecha límite no es válida.")
  ```

  (Si `go vet`/staticcheck se queja del mensaje en mayúscula/punto, usar `errors.New("fecha límite no válida")` y un `const DueDateInvalidMessage = "La fecha límite no es válida."` para el mensaje de usuario.)

- `ParseDueDate(s string) (*time.Time, error)`: `strings.TrimSpace`; cadena vacía → `nil, nil`; si no, `time.Parse(DueDateLayout, s)` (resultado en UTC a medianoche); error → `nil, ErrInvalidDueDate`.
- `(t Task) DueDateValue() string`: `""` si nil; si no `t.DueDate.Format(DueDateLayout)`. Para el `value` del input y el atributo `datetime`.
- `(t Task) DueDateLabel() string`: `""` si nil; si no `t.DueDate.Format("02/01/2006")`.

Usar receptor de valor para que Plush pueda llamarlos tanto sobre `task` (`*Task`) como sobre los elementos de `tasks` (`Task`).

### Acciones (`actions/tasks.go`)

`TasksCreate`:

- Sustituir `c.Bind(task)` por el mismo patrón que `TasksUpdate`:

  ```go
  input := struct {
  	Title   string `form:"title" json:"title"`
  	DueDate string `form:"due_date" json:"due_date"`
  }{}
  if err := c.Bind(&input); err != nil {
  	return err
  }
  task := &models.Task{Title: strings.TrimSpace(input.Title)}
  ```

- Tras calcular la posición:

  ```go
  verrs := validate.NewErrors()
  dueDate, err := models.ParseDueDate(input.DueDate)
  if err != nil {
  	verrs.Add("due_date", <mensaje>)
  } else {
  	task.DueDate = dueDate
  	verrs, err = tx.ValidateAndCreate(task)
  	if err != nil {
  		return err
  	}
  }
  ```

  La rama `if verrs.HasAny()` existente se mantiene igual (recarga tareas y renderiza `422`).

`TasksUpdate`:

- Añadir `DueDate string \`form:"due_date" json:"due_date"\`` al struct `input`.
- Parsear igual; si es inválida, añadir el error a un `validate.Errors` y renderizar `422` con `tasks/edit.plush.html` **sin** llamar a `ValidateAndUpdate`; si es válida, `task.DueDate = dueDate` y seguir el flujo actual. Campo vacío → `nil` (quita la fecha).
- Actualizar el comentario de la función: «actualiza el título y la fecha límite, conservando los demás datos de la tarea».

Rutas: sin cambios.

### Plantillas

`templates/tasks/index.plush.html`:

- Dentro de cada `<li>`, sustituir `<span class="task-title"><%= t.Title %></span>` por:

  ```html
  <span class="task-title"><%= t.Title %><%= if (t.DueDate) { %><span class="task-due">Fecha límite: <time datetime="<%= t.DueDateValue() %>"><%= t.DueDateLabel() %></time></span><% } %></span>
  ```

  (Si Plush no evalúa un puntero nil como falso, usar `<%= if (t.DueDateLabel() != "") { %>`.)

- En `.new-task-row`, entre el input de título y el botón:

  ```html
  <label for="task-due-date" class="sr-only">Fecha límite (opcional)</label>
  <input id="task-due-date" type="date" name="due_date" value="<%= task.DueDateValue() %>" <%= if (len(errors.Get("due_date")) > 0) { %>aria-invalid="true" aria-describedby="title-error"<% } %>>
  ```

- En `<p id="title-error">` iterar también `errors.Get("due_date")` (mismo bucle tras el de título). Mantener `aria-invalid` del título como está.

`templates/tasks/edit.plush.html`:

- Tras el `<p id="title-help">` (y antes del bloque de error), añadir:

  ```html
  <label for="task-due-date">Fecha límite (opcional)</label>
  <input id="task-due-date" type="date" name="due_date" value="<%= task.DueDateValue() %>" aria-describedby="due-date-help">
  <p id="due-date-help" class="field-hint">Déjala vacía para que la tarea no tenga fecha límite.</p>
  ```

- En el `<p id="title-error">` iterar también `errors.Get("due_date")`.
- Cambiar el hint del título a «Hasta 200 caracteres.» (la frase «Los demás datos de la tarea no cambian» deja de ser exacta). Comprobar que ningún test busca el texto antiguo.

### CSS (`public/assets/app.css`)

- Extender el selector `input[type="text"]` a `input[type="text"], input[type="date"]`.
- Añadir:

  ```css
  .new-task-row input[type="date"] {
    flex: 0 0 auto;
    width: auto;
  }

  .task-due {
    display: block;
    color: var(--color-muted);
    font-size: 0.8rem;
  }

  .edit-task input[type="date"] {
    margin-top: 0;
  }
  ```

  (El último solo si hace falta separación; en la edición añadir un `margin-top` a la segunda `label` si queda pegada al hint anterior, p. ej. `.edit-task label + input[type="date"]` o `.edit-task .field-hint + label { margin-top: 1.25rem; }`.)
- En `@media (max-width: 380px)`, que el input de fecha no herede `flex-basis: calc(100% - 44px)` de forma que rompa la fila: `.new-task-row input[type="date"] { flex-basis: auto; }` es aceptable (la fila ya hace wrap).
- `templates/application.plush.html`: `app.css?v=5` → `app.css?v=6`.

En tareas completadas la fecha hereda el tachado del título; se acepta (la tarea ya no está pendiente).

## Archivos relevantes

- `migrations/20261026000000_add_due_date_to_tasks.up.sql` / `.down.sql` — nuevos; columna `due_date`.
- `models/task.go` — campo `DueDate`, `DueDateLayout`, `ParseDueDate`, `DueDateValue`, `DueDateLabel`.
- `models/task_test.go` — tests de persistencia y helpers.
- `actions/tasks.go` — `TasksCreate`, `TasksUpdate`.
- `actions/tasks_test.go` — tests de creación, listado, edición y actualización con fecha.
- `templates/tasks/index.plush.html` — input de fecha y visualización.
- `templates/tasks/edit.plush.html` — input de fecha y errores.
- `public/assets/app.css` — estilos del input y de `.task-due`.
- `templates/application.plush.html` — versión del CSS.

## Implementación paso a paso

1. Crear las dos migraciones SQL.
2. Modelo: campo, constante, error/mensaje y los tres helpers.
3. Tests de modelo (paso 8) y ejecutarlos.
4. `TasksCreate`: binding con struct de entrada y parseo de la fecha.
5. `TasksUpdate`: añadir `due_date`, parseo y rama de error.
6. Plantillas `index` y `edit`.
7. CSS y bump `app.css?v=6`.
8. Tests:
   - Modelo (`models/task_test.go`):
     - `Test_Task_DueDate_Persists`: crear con `ParseDueDate("2026-11-03")`, recargar y comprobar `DueDateValue() == "2026-11-03"` y `DueDateLabel() == "03/11/2026"`.
     - `Test_Task_DueDate_Optional`: crear sin fecha, recargar, `DueDate == nil`, `DueDateValue() == ""`.
     - `Test_ParseDueDate`: `""` y `"  "` → nil sin error; `"2026-02-30"`, `"03/11/2026"`, `"2026-11-03T10:00"` → `ErrInvalidDueDate`.
   - Acciones (`actions/tasks_test.go`, usando `url.Values`):
     - `Test_TasksCreate_WithDueDate`: POST `title`+`due_date=2026-11-03` → `303`; la tarea tiene esa fecha.
     - `Test_TasksCreate_WithoutDueDate`: POST con `due_date=""` → `303`; `DueDate == nil`.
     - `Test_TasksCreate_InvalidDueDate`: → `422`, cuerpo contiene «La fecha límite no es válida.», count 0.
     - `Test_TasksIndex_ShowsDueDate`: tarea con fecha → cuerpo contiene `datetime="2026-11-03"` y `03/11/2026`; el formulario contiene `name="due_date"`.
     - `Test_TasksEdit_PrefillsDueDate`: `value="2026-11-03"`.
     - `Test_TasksUpdate_SetsAndClearsDueDate`: PUT con fecha → guardada; PUT con `due_date=""` → nil; `completed`/`position` intactos.
     - `Test_TasksUpdate_InvalidDueDate`: → `422`, mensaje visible, título y fecha sin cambios en BD.
   - Los tests existentes (`Test_TasksCreate_Valid`, `Test_TasksUpdate_PersistsTitleOnly`, etc.) deben seguir pasando sin modificarse.

## Tests y validación

Requisitos: PostgreSQL accesible según `database.yml` (entorno `test`) y migraciones aplicadas con `soda migrate -e test` (como en `.github/workflows/ci.yaml`), con `GO_ENV=test`.

- `gofmt -l .` → sin salida.
- `go vet ./...`
- `soda migrate -e test` (aplica la nueva migración) y, si es posible, `soda migrate down -e test` + `soda migrate -e test` para comprobar el down.
- `go test -p 1 ./...`
- `go build ./...`
- Manual opcional (`/env:start`): crear tarea con y sin fecha, editar, quitar fecha; revisar a 380 px de ancho.

## Riesgos y recuperación

- **Zona horaria al escribir/leer `date`**: `ParseDueDate` produce medianoche UTC y se formatea con los campos de la propia `time.Time`; con el driver pgx un parámetro de columna `date` se codifica por año/mes/día. El test de ida y vuelta (`Test_Task_DueDate_Persists`) lo verifica; si fallase por desplazamiento de día, almacenar/leer mediante `nulls.Time` o formatear en UTC explícitamente (`t.DueDate.UTC()`).
- **`*time.Time` en pop**: si pop no gestiona el puntero en insert/update, usar `nulls.Time` (`github.com/gobuffalo/nulls`, ya presente en `go.sum` como indirecta; requeriría pasarlo a directa en `go.mod`). Dejar constancia en el PR si se cambia.
- **Plush y punteros nil**: si `if (t.DueDate)` no se comporta como se espera, usar `t.DueDateLabel() != ""`.
- **Cambio de binding en `TasksCreate`**: se pasa de `c.Bind(task)` a un struct de entrada; los tests existentes de creación cubren que el título sigue llegando.
- Recuperación: revertir el commit y ejecutar `soda migrate down` (la columna es nueva y opcional; borrarla no afecta a otros datos).
