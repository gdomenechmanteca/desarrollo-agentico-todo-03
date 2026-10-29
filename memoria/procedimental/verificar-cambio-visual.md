# Verificar un cambio visual con evidencias

## Cuándo aplicarla
Cambios de CSS o plantillas sin tests automáticos (colores, tamaños, visibilidad), típicamente dentro de `/revision`.

## Pasos
1. Tener la app respondiendo en `http://localhost:3000` con la copia de trabajo (en Docker, `up --build`).
2. Escribir un script para `playwright-cli run-code --filename=<script>` **en el scratchpad**. El script debe:
   - navegar a cada página afectada (`/`, `/tasks/<id>/edit`; el id se obtiene con `curl -s / | grep -o 'href="/tasks/[^"]*/edit"'`);
   - sacar el ratón (`page.mouse.move(0,0)`), leer `getComputedStyle` del elemento y capturar la página;
   - repetir tras `page.hover(...)` y tras dar el foco con el teclado, si aplica;
   - leer también los elementos que **no** deben cambiar (p. ej. `.brand-dot`, `.add-mark`, botón secundario).
3. Lanzarlo desde la carpeta de evidencias: `playwright-cli open <url>`, `resize 1280 800`, `--raw run-code ... > <req>-estilos.json` y `close`. La salida `--raw` llega como cadena JSON escapada: decodificarla con `node -e "...JSON.parse(...)"`.
4. Para el contraste WCAG, usar `node -e` con la fórmula de luminancia relativa (no hay Python).
5. Para comprobar la caché, guardar `curl -i` del asset versionado y `grep app.css` de la página.
6. **Antes de hacer `git add`**, borrar o excluir la carpeta `.playwright-cli/` que la CLI crea en el directorio actual.

## Comprobación de que ha funcionado
El JSON muestra los `rgb(...)` esperados del plan en reposo, hover y foco, y las capturas lo confirman a simple vista (abrirlas con Read).

## Errores conocidos
- `.playwright-cli/` acabó en el commit (`979efd3` tuvo que quitarla).
- Usos: 1 (éxitos: 1) · Última vez: 2026-10-29
- Fuentes: [[episodica/2026-10-29-boton-color-verde]]
