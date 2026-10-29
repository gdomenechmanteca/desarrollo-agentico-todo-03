# Propuestas de cambio a las skills

Cambios a `.claude/skills/` y `.claude/commands/` deducidos de episodios. **No se aplican solos**: una persona los aprueba y, una vez aplicados, se marcan `aplicada` con la fecha.

## P1 · revision §3 y §Veredicto: fallos preexistentes del entorno
- Problema: el veredicto obliga a RECHAZADO si falla `gofmt` o un test. En Windows, ambos fallan siempre por CRLF sin relación con el cambio. En el episodio se resolvió con criterio propio (APROBADO CON OBSERVACIONES), contradiciendo la letra de la skill.
- Evidencia: [[episodica/2026-10-29-boton-color-verde]]
- Cambio propuesto (añadir al apartado *Veredicto*): "Un fallo de `gofmt`, `go vet` o tests no cuenta para RECHAZADO si afecta solo a archivos que el cambio no toca y se demuestra preexistente (el mismo fallo sobre la base, o la causa de entorno identificada con evidencia). Se declara como observación de entorno y debe confirmarlo el CI."
- Estado: pendiente

## P2 · revision §3: leer salidas completas y filtrar worktrees
- Problema: se leyó solo `tail -n3` de `gofmt.txt` y el informe afirmó que `gofmt` listaba 2 archivos, cuando eran 17 del proyecto y 65 de `.claude/worktrees/`.
- Evidencia: [[episodica/2026-10-29-boton-color-verde]]
- Cambio propuesto: sustituir "`gofmt -l .` (no debe listar archivos)" por "`gofmt -l $(git ls-files '*.go')` (solo archivos versionados; no debe listar ninguno tocado por el cambio). Lee siempre la salida completa de cada comando antes de resumirla."
- Estado: pendiente

## P3 · revision §4 y §5: no dejar artefactos de playwright-cli en evidencias
- Problema: `playwright-cli` crea `.playwright-cli/` en el directorio actual y entró en el commit.
- Evidencia: [[episodica/2026-10-29-boton-color-verde]]
- Cambio propuesto (añadir a §5): "No ejecutes herramientas desde `reviews/<slug>/evidencias/` o, si lo haces, borra antes del final los artefactos que no sean evidencias (p. ej. `.playwright-cli/`). Las evidencias son solo las que nombra el informe."
- Estado: pendiente

## P4 · env/start: alternativa cuando el 5432 lo ocupa otro proyecto
- Problema: el comando ordena detenerse ante un conflicto de puertos, pero hay una alternativa segura que no toca el proceso ajeno.
- Evidencia: [[episodica/2026-10-29-boton-color-verde]], [[procedimental/levantar-entorno-con-5432-ocupado]]
- Cambio propuesto (en *Comprobaciones previas*, punto 2): "Si el 5432 pertenece a otro proyecto y el 3000 está libre, usa el modo Docker completo con un override temporal fuera del repositorio que haga `ports: !reset []` en `db`, e informa de ello."
- Estado: pendiente

## P5 · awd §2.1: cambios sin commitear del propio agente
- Problema: la regla "si hay cambios sin commitear ajenos, detente" se incumplió porque los cambios eran skills recién creadas por el agente en la misma sesión. La skill no contempla ese caso.
- Evidencia: [[episodica/2026-10-29-boton-color-verde]]
- Cambio propuesto: "Si los cambios sin commitear están solo en `.claude/` o `memoria/` y no se van a añadir al commit, puedes continuar: decláralos en el PR como excluidos. Cualquier otro cambio ajeno obliga a detenerse."
- Estado: pendiente
