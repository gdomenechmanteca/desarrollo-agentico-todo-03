# Tests y validación

### Requisitos para ejecutar go test
- `actions` y `models` necesitan PostgreSQL con la base de datos `todo_test` y `GO_ENV=test`, migrada con `soda migrate -e test` (`soda` v6.4.1), igual que el CI (`.github/workflows/ci.yaml:28,51-55`).
- `adws` no usa la base de datos, pero sus tests lanzan `bash`. La imagen `golang:*-alpine` no lo trae, y con CRLF en el checkout el wrapper falla igualmente (ver [[semantica/entorno#el-checkout-windows-tiene-crlf-y-rompe-gofmt-y-scripts-bash]]).
- No hay tests automáticos de CSS ni de plantillas. Los cambios visuales se demuestran con estilos calculados y capturas ([[procedimental/verificar-cambio-visual]]).
- Estado: vigente
- Verificado: 2026-10-29
- Fuentes: [[episodica/2026-10-29-boton-color-verde]]
