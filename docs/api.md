# API de Kora

Referencia de la API HTTP que consume el frontend. Todas las respuestas que
aparecen en este documento se capturaron ejecutando los ejemplos contra el
servidor local.

| | |
|---|---|
| **URL base (local)** | `http://localhost:8080` |
| **Formato** | JSON en petición y respuesta |
| **Autenticación** | Solo el panel de personal, con `Authorization: Bearer <token>` |

---

## 1. Convenciones

### Envoltura de respuestas

Ninguna respuesta devuelve el dato pelado. Éxito y error tienen dos formas
distintas:

```jsonc
// Éxito
{ "data": <lo que sea> }

// Error
{ "error": "mensaje legible en español" }
```

Todo código de error superior a `499` se registra en el log del servidor con su
causa real, pero al cliente solo se le envía un mensaje genérico
(`"error interno del servidor"`) para no filtrar detalles de la base de datos.

### Cabeceras de respuesta

Toda respuesta JSON incluye:

```http
Content-Type: application/json
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, PUT, PATCH, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
```

El `OPTIONS` previo responde `204 No Content`.

### Autenticación

Las rutas bajo `/api/admin/` exigen cabecera `Authorization`. El formato es
exactamente `Bearer ` seguido del token:

```http
Authorization: Bearer 86cc46de-1f0b-4d6f-878b-9ce329414230
```

El token se obtiene con `POST /api/admin/login` y caduca a las 8 horas
(`TOKEN_EXPIRY_HOURS`). **Vive en memoria dentro del proceso del backend**, así
que si el contenedor se reinicia se pierde y hay que iniciar sesión otra vez.

### Fechas

Todas las fechas viajan en RFC 3339 con zona horaria, tal como las emite Go:

```json
"fecha_hora": "2026-10-01T04:23:57.870467Z"
```

---

## 2. Códigos de estado

| Código | Significado |
|---|---|
| `200 OK` | Consulta o acción exitosa |
| `201 Created` | Turno creado |
| `204 No Content` | Verificación previa de CORS |
| `400 Bad Request` | JSON inválido, campo obligatorio faltante o transición no permitida |
| `401 Unauthorized` | Credenciales o token inválidos |
| `404 Not Found` | El turno o la ruta no existen |
| `405 Method Not Allowed` | La ruta existe pero no acepta ese verbo HTTP |
| `409 Conflict` | El contacto ya tiene un turno activo en ese servicio |
| `500 Internal Server Error` | Fallo no controlado (el detalle queda en el log) |

### Mensajes de error

Estos son los mensajes exactos que devuelve la API:

| Mensaje | Cuándo |
|---|---|
| `JSON inválido` | El cuerpo no se pudo decodificar |
| `el número de contacto es requerido` | Falta `contacto` |
| `el número de contacto es requerido (mínimo 4 caracteres)` | `contacto` con menos de 4 caracteres |
| `el nombre es requerido` | Falta `nombre` |
| `debe seleccionar un servicio médico` | Falta `servicio_id` o es `<= 0` |
| `el servicio médico no existe o está inactivo` | `servicio_id` desconocido |
| `el tipo de atención no corresponde al servicio seleccionado` | `tipo_atencion_id` pertenece a otro servicio |
| `formato de fecha_hora inválido, use RFC3339` | `fecha_hora` mal formada |
| `ya tiene un turno activo en este servicio (turno CAR-001)` | Conflicto de turno activo (`409`) |
| `ingrese su número de turno o su número de contacto` | Consulta con `query` vacío |
| `no se encontró ningún turno con esos datos` | `GET /api/turnos/{numero}` sin coincidencias |
| `número de turno requerido` | Ruta `/api/turnos/` sin número |
| `turno no encontrado` | No existe el `id` en las rutas del personal |
| `ID de turno inválido` | El `id` de la URL no es un entero positivo |
| `ruta no encontrada` | Sufijo no reconocido bajo `/api/admin/turnos/` |
| `credenciales inválidas` | Correo o contraseña incorrectos (`401`) |
| `token de autorización requerido` | Falta la cabecera `Authorization` (`401`) |
| `formato de autorización inválido` | La cabecera no es `Bearer <token>` (`401`) |
| `token inválido o expirado` | Token desconocido o caducado (`401`) |
| `estado inválido` | `estado` no pertenece a la lista válida |
| `transición de estado no permitida: no se puede cambiar de ATENDIDO a EN_ATENCION` | Estado final |
| `el motivo no puede superar los 255 caracteres` | `motivo` demasiado largo |
| `el turno ya pertenece a ese servicio y tipo de atención` | Re-encolado al mismo destino |
| `número de turno requerido`, `método no permitido` | Validaciones de ruta |

---

## 3. Modelos

### Servicio

```jsonc
{
  "id": 3,
  "codigo": "CAR",                 // prefijo del número de turno
  "nombre": "Cardiología",
  "descripcion": "Evaluación cardiovascular y estudios del corazón.",
  "duracion_minutos": 20,          // base para el tiempo estimado de espera
  "activo": true,
  "orden": 3,
  "tipos_atencion": [              // opcional (omitempty)
    { "id": 10, "servicio_id": 3, "nombre": "Consulta", "orden": 1, "activo": true },
    { "id": 9,  "servicio_id": 3, "nombre": "Control", "orden": 2, "activo": true },
    { "id": 8,  "servicio_id": 3, "nombre": "Ecocardiograma", "orden": 3, "activo": true },
    { "id": 7,  "servicio_id": 3, "nombre": "Prueba de esfuerzo", "orden": 4, "activo": true }
  ]
}
```

### Turno

```jsonc
{
  "id": 2,
  "numero_turno": "CAR-001",
  "servicio_id": 3,
  "servicio_codigo": "CAR",
  "servicio_nombre": "Cardiología",
  "tipo_atencion_id": 10,
  "tipo_atencion_nombre": "Consulta",
  "cliente_contacto": "3005550011",
  "cliente_nombre": "Ana Torres",
  "cliente_documento": "CC 1020304050",
  "fecha_hora": "2026-10-01T04:23:57.870467Z",
  "expira_en": "2026-10-01T09:23:57.870467Z",
  "estado": "PENDIENTE",
  "orden_cola": 1,
  "llamado_en": null,              // se llena al pasar a EN_ATENCION
  "atendido_en": null,             // se llena al pasar a ATENDIDO
  "cancelado_en": null,            // se llena al cancelar
  "motivo_cancelacion": null,
  "creado_en": "2026-10-01T04:23:57.874444Z"
}
```

> Los campos marcados con `omitempty` **se omiten cuando están vacíos**. Si tu
> cliente lee `servicio_codigo` sin comprobar antes, usa `clave in objeto`.

### TicketTurno

Es un `Turno` más los datos de cola que ve el paciente. Solo tiene sentido en
estado `PENDIENTE`: en cualquier otro estado los tres últimos campos vienen en `0`.

```jsonc
{
  /* …todos los campos del turno… */
  "personas_antes": 0,      // cuántas personas con turno menor siguen esperando
  "posicion": 1,           // posición en la fila (personas_antes + 1)
  "total_en_cola": 1,
  "tiempo_estimado_min": 0 // personas_antes × duración del servicio
}
```

### Tablero

```jsonc
{
  "actualizado_en": "2026-10-01T04:24:28.683227918Z",
  "pendientes": 1,                 // total de PENDIENTE del sistema
  "en_atencion": [ /* Turno[] */ ],
  "historial": [ /* Turno[] — los 12 últimos llamados */ ],
  "ultimo_servicio": "Cardiología" // opcional: servicio del turno en atención
}
```

---

## 4. Endpoints públicos

No requieren autenticación.

### 4.1 `GET /api/servicios`

Devuelve el catálogo de servicios activos con sus tipos de atención. Es lo
primero que carga el formulario de turnos.

```bash
curl http://localhost:8080/api/servicios
```

**Respuesta `200`** (6 servicios sembrados):

```json
{
  "data": [
    {
      "id": 1,
      "codigo": "MED",
      "nombre": "Medicina General",
      "descripcion": "Atención médica general para adultos y niños.",
      "duracion_minutos": 15,
      "activo": true,
      "orden": 1,
      "tipos_atencion": [
        { "id": 3, "servicio_id": 1, "nombre": "Consulta general", "orden": 1, "activo": true },
        { "id": 2, "servicio_id": 1, "nombre": "Control", "orden": 2, "activo": true },
        { "id": 1, "servicio_id": 1, "nombre": "Vacunación", "orden": 3, "activo": true }
      ]
    }
  ]
}
```

| Servicio | `id` | Código | Duración | Tipos de atención (`id`) |
|---|---|---|---|---|
| Medicina General | 1 | `MED` | 15 min | Consulta general (3), Control (2), Vacunación (1) |
| Pediatría | 2 | `PED` | 15 min | Consulta pediátrica (6), Control de crecimiento (5), Vacunación (4) |
| Cardiología | 3 | `CAR` | 20 min | Consulta (10), Control (9), Ecocardiograma (8), Prueba de esfuerzo (7) |
| Odontología | 4 | `ODO` | 30 min | Consulta (14), Limpieza dental (13), Endodoncia (12), Ortodoncia (11) |
| Oftalmología | 5 | `OFT` | 20 min | Consulta (17), Valoración visual (16), Lentes (15) |
| Psicología | 6 | `PSI` | 40 min | Consulta (20), Terapia (19), Orientación (18) |

---

### 4.2 `POST /api/turnos`

Genera un turno y devuelve el ticket con la posición en la fila.

```bash
curl -X POST http://localhost:8080/api/turnos \
  -H "Content-Type: application/json" \
  -d '{
    "contacto": "3005550011",
    "nombre": "Ana Torres",
    "documento": "CC 1020304050",
    "servicio_id": 3,
    "tipo_atencion_id": 10
  }'
```

#### Cuerpo de la petición

| Campo | Tipo | Obligatorio | Descripción |
|---|---|---|---|
| `contacto` | string | **Sí** | Número de contacto, mínimo 4 caracteres |
| `nombre` | string | **Sí** | Nombre del paciente |
| `documento` | string | No | Documento de identidad |
| `servicio_id` | int | **Sí** | Servicio médico, debe existir y estar activo |
| `tipo_atencion_id` | int | No | Debe pertenecer al `servicio_id` indicado |
| `fecha_hora` | string | No | RFC 3339. Si se omite, se usa el momento actual |

**Respuesta `201`:**

```json
{
  "data": {
    "id": 2,
    "numero_turno": "CAR-001",
    "servicio_id": 3,
    "servicio_codigo": "CAR",
    "servicio_nombre": "Cardiología",
    "tipo_atencion_id": 10,
    "tipo_atencion_nombre": "Consulta",
    "cliente_contacto": "3005550011",
    "cliente_nombre": "Ana Torres",
    "cliente_documento": "CC 1020304050",
    "fecha_hora": "2026-10-01T04:23:57.870467Z",
    "expira_en": "2026-10-01T09:23:57.870467Z",
    "estado": "PENDIENTE",
    "orden_cola": 1,
    "creado_en": "2026-10-01T04:23:57.874444Z",
    "personas_antes": 0,
    "posicion": 1,
    "total_en_cola": 1,
    "tiempo_estimado_min": 0
  }
}
```

#### Errores posibles

| Código | Ejemplo de respuesta |
|---|---|
| `400` | `{"error":"debe seleccionar un servicio médico"}` |
| `400` | `{"error":"el tipo de atención no corresponde al servicio seleccionado"}` |
| `400` | `{"error":"el número de contacto es requerido (mínimo 4 caracteres)"}` |
| `409` | `{"error":"ya tiene un turno activo en este servicio (turno CAR-001)"}` |

> El `409` es a propósito: una persona no puede acumular dos turnos en el mismo
> servicio. Con el mismo contacto sí puede tener uno en Cardiología y otro en
> Odontología.

---

### 4.3 `GET /api/turnos/{numero}`

Devuelve el ticket de un turno a partir de su número visible.

```bash
curl http://localhost:8080/api/turnos/CAR-001
```

**Respuesta `200`:** el mismo objeto `TicketTurno` que devuelve la creación.

**Errores:**

| Código | Respuesta |
|---|---|
| `400` | `{"error":"número de turno requerido"}` |
| `404` | `{"error":"no se encontró ningún turno con esos datos"}` |

**Tolerancia del formato:** no hace falta escribirlo exactamente como se emitió.

| Escribes | Encuentra |
|---|---|
| `CAR-001` | `CAR-001` |
| `CAR1` / `car1` | `CAR-001` |
| `car 1` (con espacio) | `CAR-001` |

---

### 4.4 `POST /api/turnos/consultar`

Busca por **número de turno o por número de contacto**. La API decide cuál de las
dos cosas es según el texto recibido: si empieza con 2 a 4 letras seguidas de
números se trata como código de turno; cualquier otra cosa, como contacto.

```bash
# Por número de turno
curl -X POST http://localhost:8080/api/turnos/consultar \
  -H "Content-Type: application/json" \
  -d '{"query":"CAR-001"}'

# Por número de contacto
curl -X POST http://localhost:8080/api/turnos/consultar \
  -H "Content-Type: application/json" \
  -d '{"contacto":"3005550011"}'
```

#### Cuerpo de la petición

| Campo | Tipo | Obligatorio | Descripción |
|---|---|---|---|
| `query` | string | Uno de los dos | Número de turno o número de contacto |
| `contacto` | string | Uno de los dos | Alias de `query`, por compatibilidad |

#### Respuesta `200` — encontrado por número

```json
{
  "data": {
    "turno_activo": true,
    "criterio": "numero",
    "turno": { /* TicketTurno */ }
  }
}
```

#### Respuesta `200` — encontrado por contacto

```json
{
  "data": {
    "turno_activo": true,
    "criterio": "contacto",
    "turno": { /* TicketTurno — el más reciente */ },
    "turnos": [ /* TicketTurno[] — solo si tiene más de un servicio */ ]
  }
}
```

Cuando el contacto tiene varios servicios, `turno` es el más reciente y `turnos`
trae la lista completa. Ejemplo real con Pediatría y Oftalmología:

```json
{
  "data": {
    "turno_activo": true,
    "criterio": "contacto",
    "turno": { "id": 7, "numero_turno": "OFT-001", /* … */ },
    "turnos": [
      { "id": 7, "numero_turno": "OFT-001", /* … */ },
      { "id": 6, "numero_turno": "PED-001", /* … */ }
    ]
  }
}
```

#### Respuesta `200` — sin resultados

No es un error: se responde `200` con `turno_activo: false` y un mensaje.

```json
{ "data": { "turno_activo": false, "criterio": "numero", "mensaje": "no encontramos un turno con ese número" } }
```

```json
{ "data": { "turno_activo": false, "criterio": "contacto", "mensaje": "no tiene un turno activo" } }
```

**Errores:** `400` con `{"error":"ingrese su número de turno o su número de contacto"}`.

> Un turno `PENDIENTE` cuya `expira_en` ya pasó se marca `VENCIDO` durante esta
> consulta, así que el paciente nunca ve un turno vencido como vigente.

---

### 4.5 `GET /api/tablero`

Devuelve el estado actual de los llamados. Es el respaldo que se usa cuando el
flujo SSE no está disponible.

```bash
curl http://localhost:8080/api/tablero
```

**Respuesta `200`:**

```json
{
  "data": {
    "actualizado_en": "2026-10-01T04:24:28.683227918Z",
    "pendientes": 1,
    "en_atencion": [],
    "historial": [
      {
        "id": 2,
        "numero_turno": "CAR-001",
        "servicio_nombre": "Cardiología",
        "cliente_nombre": "Ana Torres",
        "estado": "ATENDIDO",
        "llamado_en": "2026-10-01T04:24:20.176193Z",
        "atendido_en": "2026-10-01T04:24:20.377565Z"
      }
    ]
  }
}
```

- `en_atencion`: turnos en estado `EN_ATENCION`, del llamado más reciente al más
  antiguo.
- `historial`: los **12 últimos turnos que recibieron un llamado**, ordenados por
  `llamado_en` descendente. Incluye los que después se atendieron o cancelaron.
- `ultimo_servicio`: aparece solo si hay alguien en atención.

---

### 4.6 `GET /api/eventos`

Flujo **Server-Sent Events** con los cambios en vivo. Es una conexión larga, no
una respuesta JSON.

```bash
curl -N http://localhost:8080/api/eventos
```

**Cabeceras:**

```http
HTTP/1.1 200 OK
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
X-Accel-Buffering: no
```

**Cuerpo del flujo:**

```
retry: 3000

event: snapshot
data: {"actualizado_en":"2026-10-01T04:24:45.755Z","pendientes":3,"en_atencion":[],"historial":[]}

event: turno_llamado
data: {"detalle":{"id":8,"numero_turno":"CAR-002","estado":"EN_ATENCION"},
       "tablero":{"pendientes":3,"en_atencion":[],"historial":[],"ultimo_servicio":"Cardiología"}}

: ping
```

Cada 20 segundos el servidor envía un comentario `": ping"` para mantener viva la
conexión. El `retry: 3000` inicial le indica al navegador que reintente a los 3
segundos si se corta.

#### Eventos

| Evento | Cuándo se emite | Forma del `data` |
|---|---|---|
| `snapshot` | Al abrir la conexión, y tras la limpieza de vencidos | `Tablero` |
| `turno_creado` | Al crear un turno | `TicketTurno` |
| `turno_llamado` | Al pasar a `EN_ATENCION` | `{ detalle, tablero }` |
| `turno_atendido` | Al pasar a `ATENDIDO` | `{ detalle, tablero }` |
| `turno_cancelado` | Al cancelar o al vencer | `{ detalle, tablero }` |
| `servicio_cambiado` | Al mover un turno a otro servicio | `{ detalle, tablero }` |
| `turno_vencido` | Reservado para el barredor automático | `{ detalle, tablero }` |

Con `detalle` viene el turno afectado y con `tablero` el estado **ya
recalculado**, de modo que el cliente solo tiene que pintar lo que recibe, sin
volver a consultar. `turno_creado` es la excepción: envía el ticket recién
creado, no un tablero.

---

## 5. Endpoints del personal

Requieren `Authorization: Bearer <token>`, excepto `login`.

```bash
# Obtén el token una vez y guárdalo en la terminal
TOKEN=$(curl -s -X POST http://localhost:8080/api/admin/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@kora.com","password":"admin123"}' \
  | python -c "import sys,json;print(json.load(sys.stdin)['data']['token'])")
```

### 5.1 `POST /api/admin/login`

Inicia sesión y devuelve un token para las demás rutas.

```bash
curl -X POST http://localhost:8080/api/admin/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@kora.com","password":"admin123"}'
```

| Campo | Tipo | Obligatorio |
|---|---|---|
| `email` | string | Sí |
| `password` | string | Sí |

**Respuesta `200`:**

```json
{ "data": { "token": "86cc46de-1f0b-4d6f-878b-9ce329414230" } }
```

**Error `401`:** `{"error":"credenciales inválidas"}`. El mensaje es el mismo
para correo inexistente, contraseña incorrecta y usuario sin rol `ADMIN`.

---

### 5.2 `GET /api/admin/turnos`

Lista turnos con filtros combinables. Devuelve objetos `Turno`, no tickets: la
posición en la fila no tiene sentido en el panel de personal.

```bash
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/admin/turnos?estado=PENDIENTE&servicio_id=5&hoy=true&limite=50"
```

| Parámetro | Tipo | Descripción |
|---|---|---|
| `estado` | string | `PENDIENTE`, `EN_ATENCION`, `ATENDIDO`, `CANCELADO` o `VENCIDO`. Coincidencia exacta |
| `servicio_id` | int | Filtra por servicio |
| `contacto` | string | Coincidencia **exacta** del número de contacto, no parcial |
| `numero` | string | Coincidencia **exacta** del número de turno, no distingue mayúsculas |
| `hoy` | bool | Solo turnos con `creado_en` en la fecha actual del servidor |
| `limite` | int | Máximo de resultados; sin límite si se omite |

**Respuesta `200`:** arreglo de `Turno` ordenado **por posición de cola ascendente**
(`orden_cola`, y a igualdad por fecha de creación). No es del más reciente al más
antiguo: es el orden de la fila, que es lo que el personal necesita ver.

```json
{
  "data": [
    {
      "id": 2,
      "numero_turno": "CAR-001",
      "servicio_codigo": "CAR",
      "servicio_nombre": "Cardiología",
      "tipo_atencion_nombre": "Consulta",
      "cliente_contacto": "3005550011",
      "cliente_nombre": "Ana Torres",
      "estado": "PENDIENTE",
      "orden_cola": 1,
      "creado_en": "2026-10-01T04:23:57.874444Z"
    }
  ]
}
```

Combinaciones reales comprobadas:

| Consulta | Resultado |
|---|---|
| `?numero=OFT-001` | 1 → `OFT-001:PENDIENTE` |
| `?numero=oFt-001` | 1 → no distingue mayúsculas |
| `?numero=OFT` | **0** → la coincidencia es exacta, no parcial |
| `?contacto=3044444444` | 2 → `PED-001`, `OFT-001` |
| `?servicio_id=5&hoy=true` | 1 → `OFT-001:PENDIENTE` |
| `?estado=PENDIENTE` | 5, ordenados por cola: `ODO-002(2)`, `PSI-002(2)`, `CAR-003(3)`, `MED-004(4)`, `CAR-005(5)` |

**Error `400`:** `{"error":"estado inválido"}`.

---

### 5.3 `GET /api/admin/turnos/pendientes`

Atajo para listar únicamente los turnos `PENDIENTE`. No acepta filtros.

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/admin/turnos/pendientes
```

**Respuesta `200`:** arreglo de `Turno` en `PENDIENTE`, también ordenado por
posición de cola.

```json
{
  "data": [
    {
      "id": 2,
      "numero_turno": "CAR-001",
      "servicio_nombre": "Cardiología",
      "cliente_nombre": "Ana Torres",
      "estado": "PENDIENTE",
      "orden_cola": 1
    }
  ]
}
```

---

### 5.4 `POST /api/admin/turnos/{id}/llamar`

Pasa el turno de `PENDIENTE` a `EN_ATENCION`. Es **idempotente**: llamarlo dos
veces devuelve `200` con el mismo turno, sin error.

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/admin/turnos/2/llamar
```

**Respuesta `200`:**

```json
{
  "data": {
    "id": 2,
    "numero_turno": "CAR-001",
    "servicio_nombre": "Cardiología",
    "cliente_nombre": "Ana Torres",
    "estado": "EN_ATENCION",
    "orden_cola": 1,
    "llamado_en": "2026-10-01T04:24:20.176193Z",
    "creado_en": "2026-10-01T04:23:57.874444Z"
  }
}
```

**Errores:**

| Código | Respuesta |
|---|---|
| `400` | `{"error":"ID de turno inválido"}` |
| `400` | `{"error":"transición de estado no permitida: no se puede cambiar de ATENDIDO a EN_ATENCION"}` |
| `404` | `{"error":"turno no encontrado"}` |

---

### 5.5 `POST /api/admin/turnos/{id}/atender`

Marca el turno como `ATENDIDO`. También es idempotente.

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/admin/turnos/2/atender
```

**Respuesta `200`:**

```json
{
  "data": {
    "id": 2,
    "numero_turno": "CAR-001",
    "estado": "ATENDIDO",
    "orden_cola": 1,
    "llamado_en": "2026-10-01T04:24:20.176193Z",
    "atendido_en": "2026-10-01T04:24:20.377565Z",
    "creado_en": "2026-10-01T04:23:57.874444Z"
  }
}
```

---

### 5.6 `POST /api/admin/turnos/{id}/cancelar`

Cancela el turno. Solo se acepta desde `PENDIENTE` o `EN_ATENCION`. El cuerpo es
opcional: si se manda sin él, el motivo queda vacío.

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"motivo":"El paciente se retiró de la sala de espera"}' \
  http://localhost:8080/api/admin/turnos/3/cancelar
```

| Campo | Tipo | Obligatorio | Límite |
|---|---|---|---|
| `motivo` | string | No | 255 caracteres |

**Respuesta `200`:**

```json
{
  "data": {
    "id": 3,
    "numero_turno": "ODO-001",
    "servicio_nombre": "Odontología",
    "cliente_nombre": "Luis Mejía",
    "estado": "CANCELADO",
    "orden_cola": 1,
    "cancelado_en": "2026-10-01T04:24:27.970081Z",
    "motivo_cancelacion": "El paciente se retiró de la sala de espera",
    "creado_en": "2026-10-01T04:24:27.821084Z"
  }
}
```

**Errores:**

| Código | Respuesta |
|---|---|
| `400` | `{"error":"el motivo no puede superar los 255 caracteres"}` |
| `400` | `{"error":"transición de estado no permitida: no se puede cancelar un turno en estado CANCELADO"}` |
| `404` | `{"error":"turno no encontrado"}` |

---

### 5.7 `PATCH /api/admin/turnos/{id}/estado`

Cambia el estado directamente, validando la transición. Sirve para los casos que
no tienen atajo propio.

```bash
curl -X PATCH -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"estado":"VENCIDO"}' \
  http://localhost:8080/api/admin/turnos/2/estado
```

| Campo | Tipo | Obligatorio |
|---|---|---|
| `estado` | string | Sí — uno de los cinco estados |

**Respuesta `200`:** el `Turno` con el estado nuevo.

**Transiciones permitidas:**

| Desde | Puede pasar a |
|---|---|
| `PENDIENTE` | `EN_ATENCION`, `CANCELADO`, `VENCIDO` |
| `EN_ATENCION` | `ATENDIDO`, `CANCELADO` |
| `ATENDIDO`, `CANCELADO`, `VENCIDO` | *(ninguno — estados finales)* |

**Errores:**

| Código | Respuesta |
|---|---|
| `400` | `{"error":"estado inválido"}` |
| `400` | `{"error":"transición de estado no permitida: no se puede cambiar de ATENDIDO a PENDIENTE"}` |

---

### 5.8 `PATCH /api/admin/turnos/{id}/servicio`

Mueve un turno `PENDIENTE` a otro servicio. **Le asigna un número y una posición
nuevos**, al final de la cola de destino; no conserva los originales.

```bash
curl -X PATCH -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"servicio_id":4,"tipo_atencion_id":14,"motivo":"Solicitó odontología"}' \
  http://localhost:8080/api/admin/turnos/4/servicio
```

| Campo | Tipo | Obligatorio | Descripción |
|---|---|---|---|
| `servicio_id` | int | **Sí** | Servicio de destino |
| `tipo_atencion_id` | int | No | Debe pertenecer al servicio de destino |
| `motivo` | string | No | Hasta 255 caracteres |

**Respuesta `200`** — nótese que el `numero_turno` pasó de `PSI-001` a `ODO-002`
y el `orden_cola` de `1` a `2`:

```json
{
  "data": {
    "id": 4,
    "numero_turno": "ODO-002",
    "servicio_id": 4,
    "servicio_codigo": "ODO",
    "servicio_nombre": "Odontología",
    "tipo_atencion_id": 14,
    "tipo_atencion_nombre": "Consulta",
    "cliente_contacto": "3022222222",
    "cliente_nombre": "Sara Núñez",
    "estado": "PENDIENTE",
    "orden_cola": 2,
    "creado_en": "2026-10-01T04:24:27.8914Z"
  }
}
```

**Errores:**

| Código | Respuesta |
|---|---|
| `400` | `{"error":"debe seleccionar un servicio médico"}` |
| `400` | `{"error":"el servicio médico no existe o está inactivo"}` |
| `400` | `{"error":"el tipo de atención no corresponde al servicio seleccionado"}` |
| `400` | `{"error":"el turno ya pertenece a ese servicio y tipo de atención"}` |
| `400` | `{"error":"transición de estado no permitida: solo se pueden re-encolar turnos PENDIENTE (estado actual EN_ATENCION)"}` |
| `409` | `{"error":"ya tiene un turno activo en este servicio (turno PED-001 en Pediatría)"}` |
| `404` | `{"error":"turno no encontrado"}` |

---

### 5.9 `POST /api/admin/turnos/vencer`

Fuerza la barredura de turnos vencidos. Es la misma operación que corre el
proceso en segundo plano cada 60 segundos; está expuesta para pruebas y para
limpiar sin esperar.

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/admin/turnos/vencer
```

**Respuesta `200`:**

```json
{ "data": { "turnos_vencidos": 0 } }
```

---

## 6. Flujo completo

```bash
API=http://localhost:8080

# 1. Cargar el catálogo para mostrar los servicios
curl $API/api/servicios

# 2. El paciente pide turno en Cardiología
curl -X POST $API/api/turnos -H "Content-Type: application/json" \
  -d '{"contacto":"3005550011","nombre":"Ana Torres","documento":"CC 1020304050","servicio_id":3,"tipo_atencion_id":10}'

# 3. El paciente consulta con el número o con el contacto
curl -X POST $API/api/turnos/consultar -H "Content-Type: application/json" -d '{"query":"CAR-001"}'
curl -X POST $API/api/turnos/consultar -H "Content-Type: application/json" -d '{"query":"3005550011"}'

# 4. El personal inicia sesión
TOKEN=$(curl -s -X POST $API/api/admin/login -H "Content-Type: application/json" \
  -d '{"email":"admin@kora.com","password":"admin123"}' \
  | python -c "import sys,json;print(json.load(sys.stdin)['data']['token'])")

# 5. Ver qué está pasando en la sala de espera
curl $API/api/tablero

# 6. Llamar y atender (id = 2 en este ejemplo)
curl -X POST -H "Authorization: Bearer $TOKEN" $API/api/admin/turnos/2/llamar
curl -X POST -H "Authorization: Bearer $TOKEN" $API/api/admin/turnos/2/atender

# 7. Reubicar a otro servicio (solo si sigue PENDIENTE)
curl -X PATCH -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"servicio_id":6,"tipo_atencion_id":19,"motivo":"Solicitó psicología"}' \
  $API/api/admin/turnos/2/servicio
```

> Los ejemplos usan `python` para leer el token. Si no lo tienes instalado,
> reemplázalo por `jq -r .data.token`.
>
> El paso 7 fallaría sobre el turno del paso 6, porque `atender` lo dejó en
> `ATENDIDO`, un estado final. El orden correcto en la vida real es reubicar
> **antes** de llamar.

---

## 7. Notas de comportamiento

- **`llamar` y `atender` son idempotentes.** Si el turno ya está en el estado
  destino, responden `200` con el turno tal cual, sin publicar un evento nuevo.
- **Buscar por contacto no es un error cuando no hay resultados.** Se responde
  `200` con `turno_activo: false` para que el frontend pueda mostrar "todavía no
  tienes turno" sin tratarlo como una falla.
- **`GET /api/turnos/{numero}` es `404`; `POST /api/turnos/consultar` es `200`
  con `turno_activo: false`.** Es intencional: la primera es una URL directa y
  la segunda es un buscador de sala de espera.
- **Los tokens son de un solo proceso.** Viven en un mapa en memoria, así que
  reiniciar el backend cierra la sesión de todo el personal.
- **Los eventos SSE no se reintentan desde el servidor.** Es el `EventSource` del
  navegador el que reconecta usando el `retry: 3000` inicial.
- **Un suscriptor lento no bloquea a los demás.** El bus de eventos descarta el
  mensaje si el canal del cliente está lleno, y ese cliente se resincroniza con
  el siguiente evento.