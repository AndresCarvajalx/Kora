# Kora

Kora es un sistema web de gestión de turnos para un centro de salud. El paciente
genera su turno, consulta su posición en la fila y sigue los llamados en vivo;
el personal llama, cancela o reubica los turnos desde un panel web.

## Funcionalidades

- **Generación de turnos** por servicio médico y tipo de atención, con nombre,
  documento de identidad y número de contacto.
- **Ticket con número visible** (`CAR-014`), posición en la fila, cuántas personas
  van antes y tiempo estimado de espera.
- **Consulta de turno** con el número de turno *o* con el número de contacto.
- **Tablero de llamados en tiempo real** (`tablero.html`) con el turno actual, los
  turnos en atención, el historial y aviso sonoro opcional.
- **Panel de personal** (`admin.html`) para llamar, atender, cancelar y cambiar de
  servicio los turnos, con filtros y búsqueda.
- **Tiempo real con Server-Sent Events**, con respaldo automático a polling.

## Tecnologías

| Capa | Tecnología |
|---|---|
| Frontend | HTML5 semántico, CSS moderno (Grid/Flexbox) y JavaScript ES6 con módulos, patrón MVC |
| Backend | Go con arquitectura en capas (handler → service → repository) |
| Base de datos | PostgreSQL |
| Despliegue | Docker Compose (nginx + Go + PostgreSQL) |

El backend usa solo tres dependencias externas: `pgx` (driver de PostgreSQL),
`uuid` (identificadores) y `golang.org/x/crypto` (bcrypt). El tiempo real se
implementa con Server-Sent Events, que ya viene en la biblioteca estándar de Go.

---

## Requisitos previos

| Recurso | Mínimo |
|---|---|
| Docker Desktop | 4.x con WSL 2 habilitado |
| Go | 1.21 o superior (probado con 1.27) |
| Git | cualquiera |
| RAM libre | 4 GB (Docker Desktop reserva 2 GB por defecto) |
| Puertos | 5432, 8080 y 8081 disponibles |

## Instalar las herramientas con winget

`winget` viene incluido en Windows 10 (2004 en adelante) y Windows 11.

```powershell
winget --version
```

### Opción A — Solo Docker Desktop (recomendada)

Con esto basta para levantar el proyecto completo: base de datos, API y frontend.

```powershell
winget install --id Docker.DockerDesktop -e --source winget
winget install --id Git.Git -e --source winget
winget install --id Microsoft.VisualStudioCode -e --source winget
```

Después de instalar Docker Desktop:

1. **Reinicia Windows.** La instalación habilita las funciones de WSL 2 y el
   servicio de Hyper-V, y no se aplican hasta que el equipo arranca de nuevo.
2. Abre Docker Desktop y espera a que diga *Docker Desktop is running*.
3. Verifica desde la terminal:
   ```powershell
   docker version
   docker compose version
   ```

### Opción B — Desarrollo sin Docker

Además de lo anterior, instala Go para compilar y ejecutar el backend en tu
máquina (el frontend se sirve con cualquier servidor estático).

```powershell
winget install --id GoLang.Go -e --source winget
```

> **Cierra y vuelve a abrir la terminal** después de instalar Go, para que el
> `PATH` se actualice. Si `go version` sigue sin funcionamando, reinicia sesión.

Opcionales, según lo que quieras usar:

```powershell
# PostgreSQL nativo, solo si no vas a usar el contenedor de base de datos
winget install --id PostgreSQL.PostgreSQL.17 -e --source winget

# PowerShell 7, cómodo para los comandos de la Opción 2
winget install --id Microsoft.PowerShell -e --source winget
```

### Herramientas de apoyo para el frontend

El frontend usa **módulos ES nativos**, así que necesita servirse por HTTP
(nunca abriendo el archivo con doble clic). Cualquiera de estas opciones sirve:

| Opción | Comando | Requisito |
|---|---|---|
| Python | `python -m http.server 8081` | Python instalado |
| Node.js | `npx --yes serve -l 8081 .` | Node.js instalado |
| VS Code | extensión *Live Server* → *Open with Live Server* | VS Code |

Si eliges Node: `winget install --id OpenJS.NodeJS.LTS -e --source winget`

---

## Puesta en marcha — Opción 1: todo con Docker

**Un solo comando levanta los tres servicios.** Este es el camino recomendado.

```bash
# 1. Clona el repositorio
git clone <url-del-repositorio> kora
cd kora

# 2. Levanta todo en segundo plano y construye las imágenes
docker compose up -d --build

# 3. Revisa que los contenedores estén arriba
docker compose ps
```

La primera compilación tarda un par de minutos (descarga Go, Alpine y nginx).
Al terminar verás:

```
NAME            IMAGE                STATUS          PORTS
kora_postgres   postgres:16-alpine   Up (healthy)    0.0.0.0:5432->5432/tcp
kora_backend    kora-backend         Up              0.0.0.0:8080->8080/tcp
kora_frontend   kora-frontend        Up              0.0.0.0:8081->80/tcp
```

### Direcciones

| Pantalla | URL |
|---|---|
| Paciente (pedir y consultar turno) | <http://localhost:8081> |
| Tablero de llamados | <http://localhost:8081/tablero.html> |
| Personal (requiere iniciar sesión) | <http://localhost:8081/admin.html> |
| API | <http://localhost:8080> |

### Credenciales iniciales

El usuario administrador viene creado por la primera migración:

| Campo | Valor |
|---|---|
| Correo | `admin@kora.com` |
| Contraseña | `admin123` |

> Esta cuenta es de ejemplo. Antes de usar el sistema en un centro de salud
> real, cambia la contraseña o crea otro usuario en la tabla `usuarios`.

### Verificar que todo responde

```bash
# El catálogo de servicios debe listar 6 servicios con sus tipos de atención
curl http://localhost:8080/api/servicios

# El estado del tablero debe venir con "pendientes" y "historial"
curl http://localhost:8080/api/tablero
```

No hace falta crear el esquema a mano: el backend aplica las migraciones
automáticamente en cada arranque y registra lo aplicado en la tabla
`schema_migrations`.

---

## Puesta en marcha — Opción 2: desarrollo sin Docker

Útil si quieres recargar el backend con `go run` y ver los logs del compilador.
La base de datos sí conviene dejarla en Docker para no instalar PostgreSQL.

### Paso 1 — Base de datos

```bash
docker compose up -d postgres
```

Esto levanta PostgreSQL en el puerto 5432 con la base `kora_turnos` y el usuario
`kora`. Espera a que el contenedor esté *healthy*.

Si prefieres PostgreSQL nativo, crea tú la base:

```sql
CREATE USER kora WITH PASSWORD 'kora123';
CREATE DATABASE kora_turnos OWNER kora;
```

### Paso 2 — Backend

Con Docker para la base de datos, apunta la API a `localhost`.

**PowerShell:**

```powershell
cd backend
$env:DATABASE_URL = "postgres://kora:kora123@localhost:5432/kora_turnos?sslmode=disable"
go run ./cmd/api
```

**Git Bash / WSL / Linux / macOS:**

```bash
cd backend
export DATABASE_URL="postgres://kora:kora123@localhost:5432/kora_turnos?sslmode=disable"
go run ./cmd/api
```

Verás en consola:

```
Conexión a base de datos establecida
Migraciones aplicadas
Servidor escuchando en :8080
```

> Si compila con un error de tipo o imports, ejecuta `go mod tidy` una vez.
> El servidor escribe `error interno: …` en consola cuando algo falla; el
> navegador solo ve un mensaje genérico.

### Paso 3 — Frontend

En **otra terminal**, desde la carpeta `frontend`:

```bash
python -m http.server 8081
```

```bash
npx --yes serve -l 8081 .
```

Luego abre <http://localhost:8081>.

> No abras `frontend/index.html` haciendo doble clic. Al usar módulos ES, el
> navegador los bloquea con el protocolo `file://` y la página se queda en
> blanco.

### Detener todo

```bash
docker compose stop        # detiene los contenedores, conserva los datos
docker compose down        # elimina los contenedores, conserva el volumen
docker compose down -v     # elimina también el volumen: BORRA LA BASE DE DATOS
```

---

## Variables de entorno

El backend lee su configuración **del entorno**, con `os.Getenv`. No carga ningún
archo `.env` automáticamente, así que los valores se definen en la terminal, en
`docker-compose.yml` o en el sistema.

`backend/.env.example` es la lista de referencia de todo lo configurable:

| Variable | Obligatoria | Por defecto | Descripción |
|---|---|---|---|
| `DATABASE_URL` | **Sí** | — | Cadena de conexión a PostgreSQL. Sin ella el backend arranca con `DATABASE_URL no está definida` |
| `PORT` | No | `8080` | Puerto en el que escucha la API |
| `TOKEN_EXPIRY_HOURS` | No | `8` | Vigencia del token del panel de personal |
| `TURNO_EXPIRY_HOURS` | No | `5` | Horas que un turno permanece vigente |

> Si prefieres trabajar con un archivo, expórtalo en la terminal antes de arrancar
> (Git Bash / WSL): `set -a && source backend/.env && set +a`

Con Docker Compose los cuatro valores ya vienen definidos en `docker-compose.yml`.
Para cambiarlos, edita ese archivo y ejecuta `docker compose up -d`.

```yaml
environment:
  DATABASE_URL: postgres://kora:kora123@postgres:5432/kora_turnos?sslmode=disable
  PORT: 8080
  TOKEN_EXPIRY_HOURS: 8
  TURNO_EXPIRY_HOURS: 5
```

---

## Comandos útiles de Docker Compose

| Comando | Qué hace |
|---|---|
| `docker compose up -d --build` | Levanta todo y reconstruye las imágenes |
| `docker compose up -d` | Levanta todo sin recompilar (si no cambiaste código) |
| `docker compose logs -f backend` | Sigue los logs del backend en vivo |
| `docker compose logs -f --tail=50` | Últimas 50 líneas de todos los servicios |
| `docker compose ps` | Estado de los contenedores |
| `docker compose restart backend` | Reinicia solo la API |
| `docker compose stop` | Detiene todo sin borrar nada |
| `docker compose down -v` | **Borra la base de datos** y empieza de cero |

Para editar código con recarga: mantén el frontend servido aparte (Opción 2) y
usa el backend con `go run`. Si cambiaste los `.sql`, no olvides reconstruir:
`docker compose up -d --build backend`.

---

## Configurar la URL de la API

El frontend resuelve solo a dónde conectarse, en este orden:

1. Si hay una URL guardada en `localStorage`, se respeta.
2. Si la página está en `localhost` o `127.0.0.1`, usa `http://localhost:8080`.
3. En cualquier otro dominio, asume que la API se sirve en el mismo origen.

Para apuntarlo manualmente a otro lugar, desde la consola del navegador:

```js
localStorage.setItem('kora_api_url', 'https://api.kora.com');
```

Y para volver al valor automático:

```js
localStorage.removeItem('kora_api_url');
```

---

## Estructura del proyecto

```
backend/
  cmd/api/main.go                       Servidor, rutas y tareas periódicas
  internal/handler/                     Capa de transporte HTTP y SSE
  internal/service/                     Reglas de negocio
  internal/repository/                  Consultas SQL
  internal/model/                       Entidades y DTOs
  internal/realtime/                    Bus de eventos en memoria
  internal/database/
    migrate.go                          Ejecutor de migraciones
    migrations/                         001_init.sql, 002_…, 003_…
frontend/
  index.html                            Vista del paciente
  tablero.html                          Pantalla de llamados en tiempo real
  admin.html                            Panel del personal
  css/styles.css                        Design system
  js/
    config.js                           URL de la API y ajustes
    utils/                              helpers de DOM, formato y audio
    models/                             communication con la API
    views/                              renderizado y manipulación del DOM
    controllers/                        orquestación de cada pantalla
docs/
  api.md                                Referencia de endpoints
  decisiones_backend.md                 Decisiones técnicas del servidor
  decisiones_frontend.md                Decisiones técnicas del cliente
docker-compose.yml                      Postgres + API + nginx
```

---

## Reglas de negocio

- El número del turno es el **consecutivo diario del servicio** (`CAR-001`,
  `CAR-002`, …), reservado de forma atómica por transacción.
- La **cola** es la combinación de servicio + tipo de atención + día. La posición
  se calcula contando los turnos `PENDIENTE` con un número de orden menor.
- Una persona puede tener **un turno activo por servicio** (índice único parcial),
  pero puede tener varios a la vez si son de servicios distintos.
- Mover un turno a otro servicio le asigna un **número y una posición nuevos** en
  la cola de destino.
- Los turnos vencen a las 5 horas de emitirse (`TURNO_EXPIRY_HOURS`) y un proceso
  en segundo plano los marca cada 60 segundos.
- Estados válidos: `PENDIENTE → EN_ATENCION → ATENDIDO`, con salidas a `CANCELADO`
  y `VENCIDO`. Los estados finales no admiten más transiciones.

---

## Solución de problemas

**`docker` no se reconoce** — Docker Desktop no está instalado o la terminal se
abrió antes de instalarlo. Cierra la terminal, vuelve a abrirla y confirma que
Docker Desktop esté corriendo en la bandeja del sistema.

**`port is already allocated` / `address already in use`** — ya hay un proceso en
ese puerto. Para el 5432 (lo más común, PostgreSQL local):

```powershell
winget install --id PostgreSQL.PostgreSQL.17 -e   # o desinstala el anterior
netstat -ano | findstr :5432                     # busca el PID
taskkill /PID <pid> /F
```

En Linux/macOS: `lsof -i :5432` y `kill <pid>`. Si prefieres mover el puerto,
cambia `5432:5432` en `docker-compose.yml`.

**El backend reinicia en bucle** — casi siempre es la base de datos. Revisa
`docker compose logs backend`; si aparece `connection refused`, PostgreSQL no
está listo todavía: `docker compose restart backend`. Si el mensaje es
`DATABASE_URL no está definida`, estás en la Opción 2 y falta exportar esa
variable antes de `go run`.

**La página se queda en blanco** — estás abriendo `index.html` con doble clic
(`file://`). Los módulos ES están bloqueados en ese protocolo. Sirve la carpeta
`frontend` por HTTP (Opción 2) o usa el contenedor de la Opción 1.

**El frontend no encuentra la API** — revisa la consola del navegador. Si hay
un error de CORS, casi siempre es que la API no está en el puerto 8080. Puedes
forzar la URL con `localStorage.setItem('kora_api_url', 'http://localhost:8080')`.

**Quiero empezar de cero** — `docker compose down -v && docker compose up -d --build`.
Esto borra el volumen de PostgreSQL y las migraciones vuelven a correr desde cero.

**`go` no se reconoce** — Go quedó instalado después de abrir la terminal. Ciérrala
y vuelve a abrirla, o reinicia sesión.

---

## Documentación

| Documento | Contenido |
|---|---|
| [docs/api.md](docs/api.md) | Todos los endpoints, sus parámetros y las respuestas reales del servidor |
| [docs/decisiones_backend.md](docs/decisiones_backend.md) | Por qué se eligió cada tecnología y cada regla del servidor |
| [docs/decisiones_frontend.md](docs/decisiones_frontend.md) | Por qué se eligió cada tecnología y cada regla del cliente |

Los directorios del proyecto siguen el patrón MVC en el frontend y arquitectura
en capas en el backend.