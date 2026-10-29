# Feature: Botón principal de color verde

## Encargo y criterios de aceptación

**Encargo literal:** "Cambia el color del boton a verde."

**Interpretación (inferida, no confirmada):** el único botón con color propio es el botón principal (`.button-primary`), que hoy es azul (`--color-button-primary: #1f5fbf`) desde el plan `plans/boton-color-azul.md`. Se usa en **"Añadir"** (`templates/tasks/index.plush.html:79`) y **"Guardar cambios"** (`templates/tasks/edit.plush.html:32`). El resto de botones (`.button` "Cancelar", iconos de acción, asa de arrastre, círculo de completar) son grises o transparentes. "El botón" se interpreta como el botón principal, igual que en el encargo anterior.

Criterios de aceptación:

1. Los botones "Añadir" y "Guardar cambios" tienen fondo y borde verdes con texto blanco. *(explícito: verde; inferido: ambos botones y texto blanco)*
2. Al pasar el ratón por encima, el fondo y el borde cambian a un verde más oscuro. *(inferido, se mantiene el patrón actual)*
3. El contraste del texto blanco sobre el verde, en reposo y en hover, cumple WCAG AA para texto normal (≥ 4.5:1). *(inferido)*
4. No cambia nada más: `--color-primary` (punto de la marca, `+` de nueva tarea, hover del círculo de completar) sigue en rojo teja, y el anillo de foco sigue azul. *(inferido: alcance mínimo)*
5. Los navegadores con la versión anterior de `app.css` en caché reciben la nueva. *(inferido, patrón del repositorio)*

## Comportamiento actual y propuesta

**Actual** (`public/assets/app.css`):

```css
--color-button-primary: #1f5fbf;
--color-button-primary-hover: #184c99;
```

`.button-primary` y `.button-primary:hover` ya usan esos tokens, y ningún otro selector los usa. `templates/application.plush.html:7` enlaza `app.css?v=6`.

**Propuesta:** cambiar solo el valor de los dos tokens:

```css
--color-button-primary: #1f7a3d;
--color-button-primary-hover: #18602f;
```

- Contraste del blanco sobre `#1f7a3d` ≈ 5,37:1; sobre `#18602f` ≈ 7,63:1 (calculado con la fórmula de luminancia relativa WCAG). Ambos cumplen AA.
- El anillo de foco azul (`--color-focus`) se dibuja por fuera del botón y sigue siendo visible.

## Alcance

**Incluido**
- Valores de `--color-button-primary` y `--color-button-primary-hover` en `public/assets/app.css`.
- Subir la versión del CSS en `templates/application.plush.html` de `app.css?v=6` a `app.css?v=7`.

**Excluido**
- Reglas `.button-primary` (ya usan los tokens), `--color-primary`, `--color-focus` y cualquier otro estilo.
- `app.js?v=4`, Go, otras plantillas, modelos, migraciones y tests.

## Diseño por capas

- **CSS** (`public/assets/app.css`): dos valores de tokens.
- **Layout** (`templates/application.plush.html`): solo la query string de versión del CSS.

## Archivos relevantes

| Ruta | Responsabilidad |
| --- | --- |
| `public/assets/app.css` | Tokens de color del botón principal (líneas 24-25) |
| `templates/application.plush.html` | Enlaza `app.css` con versión para invalidar la caché |
| `templates/tasks/index.plush.html` | Usa `.button-primary` en "Añadir" (sin cambios) |
| `templates/tasks/edit.plush.html` | Usa `.button-primary` en "Guardar cambios" (sin cambios) |

## Implementación paso a paso

1. En `public/assets/app.css`, sustituir `--color-button-primary: #1f5fbf;` por `--color-button-primary: #1f7a3d;`.
2. En `public/assets/app.css`, sustituir `--color-button-primary-hover: #184c99;` por `--color-button-primary-hover: #18602f;`.
3. En `templates/application.plush.html`, cambiar `href="/assets/app.css?v=6"` por `href="/assets/app.css?v=7"`.

## Tests y validación

No hay tests automáticos sobre CSS y no se añaden: el cambio es solo visual.

- `git diff --check`.
- `gofmt -l .` (no hay Go cambiado; comprobación de no regresión).
- `go vet ./...` y `go build ./...`: assets y plantillas se embeben (`public/embed.go`, `templates/embed.go`).
- `go test -p 1 ./...` si PostgreSQL y las migraciones de test están disponibles (no regresión).
- `grep -n "color-button-primary" public/assets/app.css`: muestra los nuevos valores.
- `grep -n "app.css?v=7" templates/application.plush.html`: devuelve una línea.
- Verificación en navegador: con la app arrancada (`/env:start`), en `/` y `/tasks/<id>/edit`, el color calculado de "Añadir" y "Guardar cambios" es `rgb(31, 122, 61)`, en hover `rgb(24, 96, 47)`, y el punto de la marca sigue `rgb(183, 53, 38)`.

## Riesgos y recuperación

- **Interpretación de "el botón":** si se refería a otro elemento o a toda la paleta, habría que cambiar también `--color-primary`. Queda anotado en el PR.
- **Tono de verde:** elegido sin especificación de diseño; se ajusta cambiando solo los dos tokens.
- **Caché:** lo cubre el paso 3.
- **Recuperación:** revertir el commit restaura el azul y la versión anterior. No hay datos ni migraciones implicadas.
