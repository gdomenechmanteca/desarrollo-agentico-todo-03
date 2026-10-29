# Levantar el entorno cuando el 5432 está ocupado por otro proyecto

## Cuándo aplicarla
`/env:start` detecta el 5432 ocupado y `docker ps` muestra que es un contenedor de **otro** proyecto (normalmente otro worktree). No hay que pararlo.

## Pasos
1. Identificar al dueño: `netstat -ano | grep ":5432 .*LISTEN"`, `tasklist //FI "PID eq <pid>"` y `docker ps --format '{{.Names}}\t{{.Ports}}\t{{.Label "com.docker.compose.project"}}'`.
2. Comprobar que el 3000 está libre.
3. Crear un override **fuera del repositorio** (en el scratchpad):
   ```yaml
   services:
     db:
       ports: !reset []
   ```
4. `docker compose -f compose.yaml -f <override> up --build -d`. La app habla con `db` por la red interna (`DB_HOST: db`) y publica solo el 3000.
5. Esperar a `docker compose ps` con ambos servicios `healthy` y a `curl http://localhost:3000/healthz` con 200.

## Comprobación de que ha funcionado
`/healthz` responde 200, `docker compose ps` muestra `5432/tcp` sin `0.0.0.0` en `db`, y el contenedor ajeno sigue igual.

## Errores conocidos
- La base de datos de este proyecto no es accesible desde el host: los tests deben ir en contenedor ([[procedimental/ejecutar-tests-en-contenedor]]) y las consultas con `docker compose exec db psql -U postgres`.
- Al terminar, el stack queda arrancado. Hay que decírselo al usuario (`docker compose stop`).
- Usos: 1 (éxitos: 1) · Última vez: 2026-10-29
- Fuentes: [[episodica/2026-10-29-boton-color-verde]]
