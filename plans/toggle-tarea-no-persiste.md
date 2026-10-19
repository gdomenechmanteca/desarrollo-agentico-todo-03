# Bug: Al marcar una tarea como completada, el cambio no persiste al recargar la página

## Encargo y resultado esperado

**Síntoma reportado:** al pulsar el botón de completar (o reabrir) una tarea, la interfaz muestra la tarea como completada, pero al recargar la página vuelve a su estado anterior.

**Resultado esperado:** al pulsar el botón, el estado `completed` de la tarea se guarda en la base de datos, de modo que después de recargar la página la tarea aparece con el estado que mostraba la interfaz. Esto vale en los dos sentidos: completar y reabrir.

## Evidencia y causa raíz

### Hechos observados en el repositorio

1. **El servidor funciona bien.** `actions/app.go` registra solo `app.PUT("/tasks/{task_id}/toggle", TasksToggle)`. `TasksToggle` (`actions/tasks.go`) invierte `task.Completed`, ejecuta `tx.Update(task)` y redirige con un 303. `Test_TasksToggle_CompletesAndReopens` (`actions/tasks_test.go:62`) confirma que un `PUT` persiste el cambio en ambos sentidos.
2. **La plantilla produce una petición correcta para el servidor.** En `templates/tasks/index.plush.html`, el formulario `data-task-toggle` usa `method="POST"` con los campos ocultos `_method=PUT` (method override de Buffalo) y `authenticity_token` (middleware `csrf.New` en `actions/app.go`).
3. **El JavaScript cancela ese envío y hace una petición incorrecta.** El commit `5fdd6c3` (*feat(ui): improve task completion feedback*) introdujo `public/assets/app.js`, que para los formularios `[data-task-toggle]`:
   - llama a `event.preventDefault()`, así que el formulario nunca llega a enviarse;
   - actualiza el DOM de forma optimista (clase `task-completed`, etiqueta, `aria-*` y la insignia "Completada");
   - termina con `fetch(form.action);`, sin opciones.
4. `fetch(url)` sin opciones hace un **GET** sin cuerpo: no envía `_method=PUT` ni `authenticity_token`. Ninguna ruta `GET /tasks/{id}/toggle` está registrada, así que la petición no llega a `TasksToggle`. El resultado de la promesa tampoco se comprueba, de modo que el fallo es silencioso y la interfaz sigue mostrando el estado optimista.

### Causa raíz

`public/assets/app.js` sustituye el envío nativo del formulario de alternar por un `fetch` que usa el método GET y omite los datos del formulario (`_method` y el token CSRF). El servidor nunca recibe la petición `PUT`, así que no se guarda nada. Al recargar, la página muestra el estado real de la base de datos.

### Incertidumbres

- No se ha comprobado en ejecución qué código devuelve exactamente `GET /tasks/{id}/toggle`. Lo más probable es un 404 (de `ServeFiles` o del router) o un 405, pero para la causa raíz basta saber que no se enruta a `TasksToggle`.
- La plantilla carga `/assets/app.js?v=2`. Si no se cambia ese parámetro, los navegadores con el script en caché podrían seguir ejecutando la versión rota después del arreglo.

### Reproducción mínima

1. Arrancar el entorno (skill `/env:start`) y abrir `/`.
2. Con una tarea pendiente, pulsar su botón de completar: la tarea se ve completada.
3. En la pestaña Network de las DevTools, ver que la petición es `GET /tasks/<id>/toggle` y que la respuesta no es un 303.
4. Recargar la página: la tarea vuelve a aparecer como pendiente.

## Alcance

**Incluido**
- Hacer que el `fetch` de `public/assets/app.js` envíe el formulario con su método real (`POST`) y con sus campos (`_method=PUT` y `authenticity_token`).
- Comprobar la respuesta: si la petición falla (respuesta no `ok` o error de red), recargar la página para que la interfaz no muestre un estado que no se ha guardado.
- Subir el parámetro de caché del script en `templates/application.plush.html` (`?v=2` → `?v=3`).

**Excluido**
- Cambios en `TasksToggle`, en las rutas, en el modelo o en las migraciones (el servidor ya funciona bien).
- Rediseñar la interacción optimista, añadir una respuesta JSON o cambiar los textos y estilos.
- Añadir infraestructura de tests de JavaScript (el repositorio no tiene ninguna: no hay `package.json` ni runner de JS).
- El manejo de `data-confirm` en el mismo listener, que no está relacionado con el bug.

## Archivos relevantes

| Ruta | Responsabilidad |
|---|---|
| `public/assets/app.js` | Intercepta el envío del formulario de alternar, actualiza el DOM y hace el `fetch`. **Origen del bug; aquí va el cambio.** |
| `templates/application.plush.html` | Carga `/assets/app.js?v=2`; hay que cambiar el parámetro de caché. |
| `templates/tasks/index.plush.html` | Formulario `data-task-toggle` con `_method=PUT` y `authenticity_token` (solo lectura, no se modifica). |
| `actions/app.go` | Ruta `PUT /tasks/{task_id}/toggle` y middleware CSRF (solo lectura). |
| `actions/tasks.go` | `TasksToggle`, que persiste el cambio (solo lectura). |
| `actions/tasks_test.go` | Tests del contrato del servidor para el toggle (se ejecutan como red de seguridad). |

## Implementación paso a paso

1. En `public/assets/app.js`, sustituir la última línea `fetch(form.action);` por un envío equivalente al del formulario nativo:

   ```js
   fetch(form.action, {
     method: form.method.toUpperCase(), // "POST"; Buffalo lee _method=PUT del cuerpo
     body: new URLSearchParams(new FormData(form)),
     credentials: "same-origin",
   })
     .then(function (response) {
       if (!response.ok) {
         throw new Error("toggle failed: " + response.status);
       }
     })
     .catch(function () {
       window.location.reload();
     });
   ```

   Notas:
   - `URLSearchParams(new FormData(form))` envía `application/x-www-form-urlencoded`, el mismo formato que el envío nativo, e incluye `_method` y `authenticity_token`.
   - `credentials: "same-origin"` garantiza que se envía la cookie de sesión de la que depende la validación CSRF (es el valor por defecto, pero se deja explícito).
   - El servidor responde 303 → `/`. `fetch` sigue la redirección y termina con un 200 de `TasksIndex`, así que `response.ok` es `true` cuando todo va bien.
   - Si algo falla, la recarga muestra el estado real de la base de datos en lugar de dejar la interfaz desincronizada.
2. Dejar intacto el resto del listener (la actualización optimista del DOM y la gestión de `data-confirm`).
3. En `templates/application.plush.html`, cambiar `/assets/app.js?v=2` por `/assets/app.js?v=3`.

## Tests y validación

**Test de regresión.** El repositorio no tiene infraestructura para ejecutar JavaScript, así que el fallo no puede cubrirse con un test automático que falle antes del arreglo sin añadir tooling, lo que queda fuera del alcance. El test de regresión es la reproducción en el navegador:

- **Antes del arreglo:** los pasos de "Reproducción mínima" muestran la petición `GET` y la tarea vuelve a pendiente al recargar (falla).
- **Después del arreglo:**
  1. Recarga forzada de `/` para obtener `app.js?v=3`.
  2. Completar una tarea: en Network aparece `POST /tasks/<id>/toggle` con `_method=PUT` y `authenticity_token` en el cuerpo, una respuesta 303 y un `GET /` con 200.
  3. Recargar: la tarea sigue completada.
  4. Reabrirla y recargar: sigue pendiente.
  5. Opcional: con el servidor parado, pulsar el botón; la página intenta recargarse y no queda un estado optimista falso.

**Comprobaciones automáticas pertinentes.** Se requieren PostgreSQL y las migraciones de test, según `database.yml` y `.github/workflows/ci.yaml`.

```bash
go test -p 1 ./actions/...
```

Los tests `Test_TasksToggle_*` deben seguir pasando: confirman que el contrato del servidor que usa el nuevo `fetch` no ha cambiado.

No hace falta ejecutar `gofmt`, `go vet` ni `go build` por código Go, porque no se toca ningún archivo `.go`. Sí conviene `go build ./...` para confirmar que las plantillas y los assets embebidos siguen compilando.

## Riesgos y recuperación

- **CSRF:** si el token no llegara en el cuerpo, el servidor respondería con un error. El `.catch` recarga la página y no se pierde ningún dato. Se detecta en el paso 2 de la validación.
- **Caché del navegador:** si no se sube `?v=`, algunos usuarios seguirían con el script roto. Por eso el paso 3 es obligatorio.
- **Doble clic rápido:** dos peticiones seguidas alternan dos veces, igual que con el envío nativo. No es una regresión y queda fuera del alcance.
- **Recuperación:** el cambio se limita a dos archivos estáticos y no toca datos ni el esquema. Se revierte con `git revert` del commit.
