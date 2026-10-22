# Feature: Botón principal de color azul

## Encargo y criterios de aceptación

**Encargo literal:** "Cambia el color del boton a azul."

**Interpretación (inferida, no confirmada):** la app solo tiene un tipo de botón con color propio, el botón principal (`.button-primary`), hoy rojo teja (`--color-primary: #b73526`). Se usa en dos sitios: **"Añadir"** (`templates/tasks/index.plush.html:77`) y **"Guardar cambios"** (`templates/tasks/edit.plush.html:28`). Los demás botones (`.button` "Cancelar", `.icon-action`, `.drag-handle`, `.task-toggle-button`) son grises o transparentes y no tienen un color que cambiar. Por eso, "el botón" se interpreta como el botón principal.

Criterios de aceptación:

1. Los botones "Añadir" y "Guardar cambios" tienen fondo y borde azules con texto blanco.
2. Al pasar el ratón por encima, el fondo y el borde cambian a un azul más oscuro.
3. El contraste del texto blanco sobre el azul, en reposo y en hover, cumple WCAG AA para texto normal (≥ 4.5:1).
4. El resto de elementos que usan `--color-primary` no cambia: el punto de la marca, el `+` de nueva tarea y el hover del círculo de completar siguen en rojo teja.
5. Los navegadores con la versión anterior de `app.css` en caché reciben la nueva.

## Comportamiento actual y propuesta

**Actual** (`public/assets/app.css`):

```css
:root {
  --color-primary: #b73526;
  --color-primary-hover: #992b1e;
  --color-focus: #245ab4;
  ...
}
.button-primary { background: var(--color-primary); border-color: var(--color-primary); color: #ffffff; }
.button-primary:hover { background: var(--color-primary-hover); border-color: var(--color-primary-hover); }
```

`--color-primary` también se usa en `.brand-dot`, `.add-mark` y `.task-toggle-button:hover .completion-icon`. Si se cambiara esa variable, toda la identidad de color pasaría a azul, más allá de lo que pide el encargo.

**Propuesta:** añadir dos tokens propios del botón y que `.button-primary` los use, sin tocar `--color-primary`:

```css
--color-button-primary: #1f5fbf;
--color-button-primary-hover: #184c99;
```

- Contraste del blanco sobre `#1f5fbf` ≈ 6,2:1; sobre `#184c99` ≈ 8,4:1. Ambos cumplen AA.
- El color se distingue del anillo de foco (`--color-focus: #245ab4`, outline de 3px con 3px de separación). El foco sigue siendo visible porque se dibuja por fuera del botón, sobre el fondo blanco.

## Alcance

**Incluido**
- Tokens nuevos y reglas `.button-primary` / `.button-primary:hover` en `public/assets/app.css`.
- Subir la versión del CSS en `templates/application.plush.html` de `app.css?v=4` a `app.css?v=5` (es el mismo patrón de cache-busting de los commits anteriores).

**Excluido**
- `--color-primary` y los elementos que lo usan (marca, `+`, hover de completar).
- El botón secundario `.button`, los iconos de acción, el asa de arrastre y el estilo de los enlaces.
- Cambios en Go, plantillas (salvo la versión del CSS), JavaScript, modelos o migraciones.

## Diseño por capas

- **CSS** (`public/assets/app.css`): dos tokens nuevos en `:root`, y `.button-primary` y `.button-primary:hover` pasan a usarlos.
- **Layout** (`templates/application.plush.html`): solo la query string de versión del CSS.

No hay cambios en persistencia, modelo, acciones, rutas ni JavaScript.

## Archivos relevantes

| Ruta | Responsabilidad |
| --- | --- |
| `public/assets/app.css` | Tokens de color y estilos de `.button-primary` |
| `templates/application.plush.html` | Enlaza `app.css` con versión para invalidar la caché |
| `templates/tasks/index.plush.html` | Usa `.button-primary` en "Añadir" (sin cambios) |
| `templates/tasks/edit.plush.html` | Usa `.button-primary` en "Guardar cambios" (sin cambios) |

## Implementación paso a paso

1. En `public/assets/app.css`, dentro de `:root`, justo después de `--color-primary-hover: #992b1e;`, añadir:
   ```css
   --color-button-primary: #1f5fbf;
   --color-button-primary-hover: #184c99;
   ```
2. En `public/assets/app.css`, sustituir las reglas del botón principal por:
   ```css
   .button-primary {
     background: var(--color-button-primary);
     border-color: var(--color-button-primary);
     color: #ffffff;
   }

   .button-primary:hover {
     background: var(--color-button-primary-hover);
     border-color: var(--color-button-primary-hover);
   }
   ```
3. En `templates/application.plush.html`, cambiar `href="/assets/app.css?v=4"` por `href="/assets/app.css?v=5"`. No tocar `app.js?v=4`.

## Tests y validación

No hay tests automáticos sobre CSS en el repositorio y no hace falta añadirlos: el cambio es solo visual.

- `git diff --check`: sin errores de espacios.
- `go build ./...`: los assets y las plantillas se embeben (`public/embed.go`, `templates/embed.go`), así que se compila para confirmar que siguen embebiéndose.
- `go vet ./...`.
- `go test -p 1 ./...` si PostgreSQL y las migraciones de test están disponibles. Ningún test depende del CSS; sirve como prueba de no regresión.
- `grep -n "button-primary\|color-button-primary" public/assets/app.css`: los tokens nuevos existen y `.button-primary` los usa.
- `grep -n "app.css?v=5" templates/application.plush.html`: devuelve una línea.
- Verificación manual (opcional): arrancar con `/env:start`, abrir `/` y `/tasks/<id>/edit`, y comprobar que "Añadir" y "Guardar cambios" son azules con hover más oscuro, y que el punto de la marca y el `+` siguen en rojo.

## Riesgos y recuperación

- **Interpretación de "el botón":** si se refería a otro elemento (por ejemplo, a toda la paleta principal), el cambio se amplía cambiando `--color-primary` y `--color-primary-hover`. Queda anotado como decisión en el PR.
- **Tono de azul:** se ha elegido sin especificación de diseño. Se ajusta cambiando solo los dos tokens.
- **Caché:** si no se sube la versión, algunos usuarios seguirían viendo el rojo. Lo cubre el paso 3.
- **Recuperación:** revertir el commit restaura los colores y la versión anteriores. No hay datos ni migraciones implicadas.
