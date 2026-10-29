# Interfaz (CSS y plantillas)

### El color del botón principal lo controlan dos tokens propios
`--color-button-primary` y `--color-button-primary-hover` (`public/assets/app.css:24-25`) solo los usa `.button-primary` (`app.css:347-356`), presente en "Añadir" (`templates/tasks/index.plush.html:79`) y "Guardar cambios" (`templates/tasks/edit.plush.html:32`). Son independientes de `--color-primary` (rojo teja `#b73526`), que colorea `.brand-dot` (`app.css:68`), `.add-mark` (`app.css:270`) y el hover de completar (`app.css:179`). Para cambiar el color del botón basta con los dos tokens; el anillo de foco usa `--color-focus` (`app.css:361`).
- Estado: vigente
- Verificado: 2026-10-29
- Fuentes: [[episodica/2026-10-29-boton-color-verde]]

### Cualquier cambio en app.css exige subir ?v=N en el layout
`templates/application.plush.html:7` enlaza `app.css?v=N` (7 tras el botón verde) y `:30` enlaza `app.js?v=N` (4), con contadores independientes. Los assets se sirven con `Cache-Control: max-age=31536000`, así que sin subir la versión los navegadores siguen usando el CSS anterior durante un año.
- Estado: vigente
- Verificado: 2026-10-29
- Fuentes: [[episodica/2026-10-29-boton-color-verde]] (`reviews/boton-color-verde/evidencias/R5-curl.txt`)
