---
description: Start the TODO development environment and verify that the application responds
allowed-tools: Bash, Read
---

# Arranca el entorno de desarrollo

Levanta la aplicación desde la raíz del repositorio. No mates procesos existentes ni elimines volúmenes.

## Modos

- Usa el modo **host** por defecto.
- Usa el modo **Docker completo** si `$ARGUMENTS` contiene `--docker` o si Go/Buffalo no están disponibles en el host.

## Comprobaciones previas

1. Confirma que `docker compose` está disponible y que el daemon responde.
2. Comprueba si los puertos `3000` o `5432` ya están ocupados.
   - Si pertenecen a este proyecto y el servicio responde, reutilízalos.
   - Si el `5432` pertenece a otro proyecto (compruébalo con `docker ps`) y el `3000` está libre, no lo finalices. Usa el modo **Docker completo** con un override temporal **fuera del repositorio**, que no publique el puerto de la base de datos, e informa de ello:
     ```yaml
     services:
       db:
         ports: !reset []
     ```
     Arranca con `docker compose -f compose.yaml -f <override> up --build -d`. La base de datos solo será accesible desde la red de compose: los tests van en contenedor y las consultas, con `docker compose exec db psql -U postgres`.
   - En cualquier otro conflicto con otro proceso, informa y detente. No lo finalices.

## Modo host

1. Comprueba que `go` y `buffalo` están disponibles.
   - Si falta alguno, explica el bloqueo y continúa mediante Docker completo.
2. Ejecuta `docker compose up -d db`.
3. Espera a que PostgreSQL figure como saludable en `docker compose ps`.
4. Ejecuta `go mod download`.
5. Inicia `buffalo dev` en segundo plano desde la raíz y conserva su salida para poder diagnosticar el arranque.
6. Espera hasta que `http://localhost:3000/healthz` responda correctamente.

## Modo Docker completo

1. Ejecuta `docker compose up --build -d`.
2. Espera a que `db` y `app` figuren como saludables en `docker compose ps`.
3. Comprueba `http://localhost:3000/healthz`.

## Resultado

Informa de:

- modo utilizado;
- servicios iniciados;
- URL de la aplicación;
- respuesta de `/healthz`;
- cualquier fallback, advertencia o proceso que deba mantenerse activo.
