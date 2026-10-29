# Ejecutar go test en un contenedor contra la base de datos de compose

## Cuándo aplicarla
La base de datos de este proyecto no está publicada en el host (ver [[procedimental/levantar-entorno-con-5432-ocupado]]) o se quiere reproducir el CI sin instalar `soda` en el host.

## Pasos
1. Crear `todo_test` si no existe:
   `docker compose exec -T db psql -U postgres -tc "SELECT 1 FROM pg_database WHERE datname='todo_test'" | grep -q 1 || docker compose exec -T db createdb -U postgres todo_test`
2. Lanzar desde Git Bash, con `MSYS_NO_PATHCONV=1` para que no reescriba las rutas:
   ```bash
   MSYS_NO_PATHCONV=1 docker run --rm --network desarrollo-agentico-todo-03_default \
     -e GO_ENV=test -e DB_HOST=db \
     -v "C:/desarrollo/ENTAINA/desarrollo-agentico-todo-03:/src" -v todo-gomodcache:/go/pkg/mod -w /src \
     golang:1.26-alpine sh -c 'apk add -q bash git && go install github.com/gobuffalo/pop/v6/soda@v6.4.1 && soda migrate -e test && go test -p 1 ./...'
   ```
3. Guardar la salida completa como evidencia y leerla **entera**, no solo el final.

## Comprobación de que ha funcionado
`ok todo/actions` y `ok todo/models`.

## Errores conocidos
- Sin `apk add bash`, todos los tests de `adws` fallan con `exit status 127`.
- Con el checkout CRLF de Windows, `adws/TestTriggerShellWrapper` falla igualmente (`pipefail: invalid option name`). Es un fallo de entorno, no del cambio; lo confirma el CI.
- Usos: 1 (éxitos: 1, con el fallo de entorno de `adws` declarado) · Última vez: 2026-10-29
- Fuentes: [[episodica/2026-10-29-boton-color-verde]]
