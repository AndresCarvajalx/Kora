# Decisiones del Proyecto — Backend

A continuación se encuentran las decisiones técnicas tomadas para el desarrollo
de la API y la lógica del lado del servidor, junto con la justificación de cada
una y sus consecuencias sobre el resto del sistema.

---

## 1. Lenguaje de programación

Se eligió **Go** como lenguaje principal para el desarrollo de la API REST del
backend.

### Principales beneficios

- **Alto rendimiento y concurrencia:** maneja eficientemente múltiples peticiones
  simultáneas mediante *goroutines*, lo cual es ideal para un sistema de gestión
  de turnos en tiempo real con bajo consumo de recursos.
- **Compilado y tipado estático:** captura errores de tipos en tiempo de
  compilación, lo que garantiza un código más robusto, predecible y menos propenso
  a fallos en producción.
- **Simplicidad y curva de aprendizaje:** cuenta con una sintaxis limpia y directa,
  lo que facilita el mantenimiento, las revisiones de código (*code reviews*) y la
  rapidez en la entrega del MVP.
- **Cero dependencias pesadas:** genera un único archivo ejecutable binario, lo que
  simplifica drásticamente el proceso de despliegue en entornos como contenedores
  o servicios serverless.

---

## 2. Arquitectura en capas

Se adoptó una arquitectura de **capas** (no MVC), adecuada para una API donde no
existe la noción de "vista".

```
handler  →  service  →  repository  →  PostgreSQL
```

- **Handlers (capa de transporte):** reciben la petición HTTP, validan el formato
  JSON de entrada, parsean los parámetros de la URL, traducen los errores de la
  capa de negocio a códigos de estado y emiten la respuesta.
- **Services (capa de negocio):** contienen las reglas centrales del sistema de
  turnos: validación de la entrada, transiciones de estado, cálculo de la posición
  en la fila, caducidad y publicación de eventos en tiempo real.
- **Repositories (capa de datos):** realizan exclusivamente las consultas SQL y se
  encargan de traducir filas a entidades.
- **Models (entidades):** definen las estructuras de Go (`structs`) que representan
  el negocio, incluyendo los DTOs que viajan por la red.

### Principales beneficios

- **Aislamiento de la lógica:** si se cambia el motor de base de datos, solo se
  modifica `repository`; la regla de negocio y los controladores quedan intactos.
- **Una sola dirección de dependencia:** los handlers nunca escriben SQL y los
  repositorios nunca toman decisiones de negocio.
- **Testabilidad:** permite probar las reglas de turnos con *mocks* de la base de
  datos sin levantar PostgreSQL.
- **Escalabilidad gradual:** evita archivos gigantescos con múltiples
  responsabilidades a medida que el MVP crece.

---

## 3. Base de datos relacional (PostgreSQL)

Se eligió **PostgreSQL** como sistema de gestión de bases de datos relacionales.

### Principales beneficios

- **ACID estricto:** garantiza transacciones seguras, lo que resulta indispensable
  al reservar números de turno de forma concurrente.
- **Relaciones claras:** los turnos dependen de `servicios` y `tipos_atencion`
  mediante claves foráneas, un modelo que se adapta de forma natural a SQL.
- **Índices parciales y expressions únicas:** permiten expresar reglas como *"un
  turno activo por persona y servicio"* directamente en el motor de la base, en
  lugar de dejarlas solo en el código.
- **Compatibilidad con proveedores cloud:** se conecta a Neon, Supabase o
  PostgreSQL nativo únicamente con una variable de entorno estándar.

### Esquema

| Tabla | Contenido |
|---|---|
| `servicios` | Catálogo médico. Su `codigo` es el prefijo del número de turno |
| `tipos_atencion` | Tipos de atención, **pertenecientes a un servicio** |
| `turnos` | Turnos emitidos, con su estado y su posición en la cola |
| `turno_secuencia` | Contador diario por servicio, para la numeración atómica |
| `usuarios` | Credenciales del personal (`password_hash` con bcrypt) |
| `schema_migrations` | Registro de las migraciones ya aplicadas |

---

## 4. Gestión de configuración y seguridad

- **Variables de entorno:** todas las credenciales y parámetros (cadena de
  conexión, puerto, vigencias) se leen del entorno, cumpliendo el principio
  *twelve-factor app* y evitando exponer contraseñas en el repositorio. El
  ejemplo de configuración vive en `backend/.env.example`.
- **Contraseñas con hash:** se usa **bcrypt** (`golang.org/x/crypto`), nunca
  texto plano ni un hash rápido, porque las contraseñas del personal son
  credenciales reutilizables.

### 4.1 Autenticación del personal con tokens en memoria

El panel de personal se autentica con un token opaco (UUID v4) que se entrega al
iniciar sesión y se envía después como `Authorization: Bearer <token>`. El token
se valida contra un mapa en memoria protegido con `sync.RWMutex`, porque varias
peticiones concurrentes del panel leen y escriben el mismo mapa a la vez.

**Limitaciones conocidas y aceptadas en el MVP:**

- Los tokens **se pierden al reiniciar el backend**, lo que cierra la sesión de
  todo el personal.
- No sirven para réplicas horizontales: cada proceso tendría su propio mapa.
- La ruta de mejora es Redis o una tabla de sesiones en PostgreSQL.

Se eligió así porque evita introducing una dependencia de infraestructura nueva
en un proyecto de alcance académico, y el impacto real (una sesión abierta) es
compatible con el uso real del panel.

### 4.2 CORS

El middleware `CORS` responde a todas las peticiones con:

```http
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, PUT, PATCH, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
```

Se optó deliberadamente por el origen abierto (`*`) porque el frontend se sirve
como archivos estáticos y, en desarrollo, puede venir de `localhost:8081` o de
cualquier puerto equivalente. **Es una limitación conocida**: en un despliegue
real habría que restringir `Allow-Origin` a los dominios autorizados y añadir
cookies `HttpOnly` para el token.

---

## 5. Migraciones automáticas

El esquema se versiona en `internal/database/migrations/*.sql` y se aplica **al
arrancar el servidor**, dejando registro en la tabla `schema_migrations`.

### Principales beneficios

- **Un solo camino de despliegue:** no depende de scripts manuales ni de que el
  volumen de Docker se inicialice por primera vez.
- **Idempotencia:** cada migración se ejecuta una sola vez y queda registrada con
  su versión; las nuevas se aplican en orden.
- **Portable:** al usar `//go:embed`, los `.sql` viajan dentro del binario y no
  pueden perderse al construir la imagen.
- **Evolución segura:** lasSentencias usan `IF NOT EXISTS` y `IF EXISTS`, por lo
  que también son aplicables sobre bases creadas por versiones anteriores.

| Migración | Contenido |
|---|---|
| `001_init.sql` | Tablas base, usuario administrador inicial y datos de prueba |
| `002_turnos_servicios.sql` | Servicios, tipos de atención, `numero_turno`, `orden_cola` y backfill de filas heredadas |
| `003_catalogo_servicios.sql` | Catálogo real de 6 servicios con sus tipos de atención |

---

## 6. Numeración de turnos y definición de la cola

El número visible del turno es el **consecutivo diario por servicio**
(`CAR-001`, `CAR-002`, …), donde el prefijo es el `codigo` del servicio. Se
reserva con un `INSERT … ON CONFLICT DO UPDATE … RETURNING` sobre la tabla
`turno_secuencia`, dentro de la misma transacción que crea el turno.

La **cola** se define por la combinación de `servicio_id` + `tipo_atencion_id` +
`cola_fecha`. La posición real se calcula contando los turnos `PENDIENTE` de esa
misma cola con un `orden_cola` menor.

### Principales beneficios

- **Atomicidad sin bloqueos explícitos:** es una sola sentencia de PostgreSQL, por
  lo que dos pacientes que generan turno en el mismo instante nunca reciben el
  mismo número.
- **Orden estable:** la posición se apoya en `orden_cola`, un valor inmutable
  asignado al crear el turno. Llamar o cancelar a otra persona nunca altera el
  orden de quienes faltan.
- **Colas realmente independientes:** una consulta y un ecocardiograma en
  Cardiología no compiten entre sí.
- **Estimación de espera sin motor de reglas:** multiplicando las personas que van
  antes por la duración del servicio se obtiene el tiempo estimado.
- **Un turno activo por servicio:** un índice único parcial sobre
  `(cliente_contacto, servicio_id)` permite tener Cardiología y Odontología a la
  vez, pero no dos turnos en el mismo servicio.

### 6.1 Normalización del número de turno

Al consultar, el backend decide si el texto recibido es un **código de turno** o
un **número de contacto** mediante la expresión regular `^[A-Z]{2,4}-?\d{1,5}$`.
El requisito de **letras iniciales** es lo que evita la ambigüedad: un contacto
solo tiene dígitos, así que nunca se confunde con un código.

Una vez identificado como código, se generan variantes de búsqueda para tolerar
que el paciente escriba `car1`, `CAR 1` o `car-01` en lugar de `CAR-001`. Es una
decisión de ergonomía: el número se muestra en pantalla y en papel, y las
personas lo transcriben a mano.

---

## 7. Estados y ciclo de vida del turno

```
PENDIENTE ──▶ EN_ATENCION ──▶ ATENDIDO
     │             │
     └──▶ CANCELADO ◀┘
     └──▶ VENCIDO
```

- Las transiciones válidas están centralizadas en `esTransicionValida`, de modo que
  ningún handler puede saltarse la máquina de estados.
- Los tres estados finales (`ATENDIDO`, `CANCELADO`, `VENCIDO`) son terminales.
- Un `ticker` cada 60 segundos marca como `VENCIDO` los turnos cuya `expira_en` ya
  pasó (5 horas por defecto, `TURNO_EXPIRY_HOURS`), y publica el cambio.
- Además del barredor, la creación y la consulta verifican la caducidad **en el
  momento**, para que un paciente nunca vea como vigente un turno vencido aunque
  el proceso todavía no haya corrido.

### 7.1 Re-encolado: cambiar de servicio asigna número nuevo

Cuando el personal mueve un turno a otro servicio, `ReencolarTurno` **no conserva
el número ni la posición originales**: el turno entra al final de la cola de
destino y recibe un número nuevo de esa cola.

Se decidió así porque:

- El número identifica la posición dentro de una cola concreta. Mantener `PSI-001`
  dentro de una cola de Odontología sería un número engañoso, tanto para el
  paciente como para el tablero de llamados.
- Reutilizar el `orden_cola` original haría que alguien que llega después se
 saltara a quien ya estaba esperando.

El turno original conserva su `id`, su contacto y su fecha de emisión, de modo que
el historial sigue siendo trazable. Solo se aceptan turnos `PENDIENTE`, y se
valida que el destino no sea el mismo servicio y tipo de atención actuales, ni uno
donde la persona ya tenga un turno activo.

---

## 8. Catálogo de servicios y tipos de atención por servicio

Los **tipos de atención pertenecen a un servicio** (`tipos_atencion.servicio_id`)
en lugar de ser un catálogo global.

La razón es que la duración y el significado de "Control" cambian según el
especialidad: un control de crecimiento en Pediatría no es lo mismo que un
control en Cardiología. Modelarlos por servicio evita ambigüedad y permite que
cada uno tenga su propia duración.

La consecuencia práctica es que **el formulario pide primero el servicio y luego
carga sus tipos de atención**. El backend valida esta correspondencia en
`CrearTurno` y en `CambiarServicioTurno`, rechazando con `400` cualquier tipo de
atención que no pertenezca al servicio elegido.

El catálogo de 6 servicios (MED, PED, CAR, ODO, OFT, PSI) viene sembrado en la
migración `003`, no en código, de modo que agregar un servicio es un `INSERT` y no
un despliegue.

---

## 9. Tiempo real con Server-Sent Events (SSE)

Los paneles se mantienen actualizados con un flujo `text/event-stream` en
`/api/eventos`. La capa de servicio publica en un **bus de eventos en memoria**
(`internal/realtime`) cada vez que cambia un turno, y cada suscriptor recibe el
evento con el estado del tablero **ya calculado**.

### Principales beneficios

- **Unidireccional y simple:** el servidor solo empuja al cliente. No hace falta
  WebSocket ni bibliotecas de terceros, lo que mantiene el compromiso de cero
  dependencias pesadas.
- **Reconexión automática:** el `retry: 3000` del protocolo hace que el navegador
  reconecte solo ante cortes de red, sin código adicional.
- **Sin compatibilidad especial:** al usar el `EventSource` estándar no se
  complican las cabeceras ni la autenticación.
- **Respaldo degradado:** si el flujo no está disponible, el modelo del frontend
  cae automáticamente a polling de `/api/tablero`.
- **El cliente no necesita recalcular nada:** cada evento trae el tablero
  completo junto con el turno afectado.

### 9.1 El bus de eventos

`internal/realtime` es un `Hub` en memoria con un mapa de canales suscriptor y un
`RWMutex`. Cada suscriptor tiene un canal con *buffer* (16 mensajes) y la
publicación es **no bloqueante**: si un cliente está saturado, el evento se
descarta (`slow client`) en lugar de frenar al resto de paneles.

Como parte de la API del paquete se expone `CantidadSuscriptores()`, pensado para
diagnóstico.

---

## 10. Envoltura de respuestas y manejo de errores

Todas las respuestas siguen la misma forma, definida en `middleware.go`:

```jsonc
{ "data": ... }     // éxito
{ "error": "..." }  // fallo
```

Cada `handler` traduce los errores tipados del service (`Err…`) al código HTTP
correspondiente: validaciones a `400`, token a `401`, ausencia de registro a `404`
y conflicto de turno activo a `409`.

La función `ErrorInterno` registra el error real con su causa en el log del
servidor, pero responde únicamente `{"error":"error interno del servidor"}`. Así
el cliente nunca ve detalles de SQL, nombres de tabla ni de columna.

---

## 11. Dependencias mínimas

El backend usa solo tres dependencias directas:

| Dependencia | Para qué |
|---|---|
| `github.com/jackc/pgx/v5` | Driver y pool de PostgreSQL |
| `github.com/google/uuid` | Identificadores de token |
| `golang.org/x/crypto` | Hash bcrypt de contraseñas |

Todo lo demás es biblioteca estándar: `net/http` con `ServeMux`, `encoding/json`
para el JSON, `log` para el logging y SSE construido sobre `http.Flusher`.

La razón es que cada dependencia es superficie de mantenimiento y de superficie de
seguridad. En un proyecto de alcance académico, reducirla a lo estrictamente
necesario hace el sistema más fácil de auditar y de desplegar.

---

## 12. Build y despliegue

- **Imagen multi-etapa:** el `Dockerfile` compila con `golang:1.21-alpine` y copia
  únicamente el binario a una imagen `alpine` final. El resultado no tiene
  toolchain ni sistema de archivos de desarrollo.
- **Binario estático:** `CGO_ENABLED=0 GOOS=linux` produce un ejecutable que corre
  en la imagen mínima sin bibliotecas de C.
- **Apagado ordenado:** se atiende `SIGINT`/`SIGTERM` y se llama a
  `srv.Shutdown` con un contexto de 5 segundos, para que las conexiones abiertas
  —incluidas las de SSE— cierren limpiamente.
- **Sin `WriteTimeout`:** el servidor deja ese campo vacío a propósito. Cualquier
  plazo de escritura cerraría las conexiones largas de SSE que son justamente las
  que mantienen vivo el tablero.
- **`ReadHeaderTimeout: 10s`:** sí se define, para evitar que un cliente lento
  agote la conexión antes de mandar sus cabeceras.