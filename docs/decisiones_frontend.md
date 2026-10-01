# Decisiones del Proyecto — Frontend

A continuación se encuentran las decisiones tomadas por el equipo para desarrollar
la interfaz de Kora, y el porqué de cada una.

Las reglas que guían este trabajo están en [`frontend/skill/skill.md`](../frontend/skill/skill.md),
que es de lectura obligatoria antes de tocar cualquier archivo del cliente.

---

## 1. Arquitectura: MVC por módulo

Se eligió la arquitectura **Modelo Vista Controlador (MVC)**, organizando el
código en módulos ES6 agrupados por **pantalla** en vez de por capa.

```
frontend/js/
  models/        adminModel, authModel, servicioModel, tableroModel, turnoModel
  views/         adminView, tableroView, turnoView
  controllers/   adminController, tableroController, turnoController
  utils/         audio, dom, format
  config.js      configuración global
  main.js        punto de entrada de index.html
  tablero.js     punto de entrada de tablero.html
  admin.js       punto de entrada de admin.html
```

- **Modelo:** maneja el estado, la lógica de datos y las peticiones `fetch` a la
  API. No toca el DOM ni conoce el HTML.
- **Vista:** renderiza el DOM, lee formularios y captura eventos. Es el único
  módulo que puede escribir HTML.
- **Controlador:** intermedia. Recibe los eventos de la Vista, invoca al Modelo y
  le indica a la Vista qué actualizar.

### Por qué agrupar por pantalla y no por capa

Un directorio `views/` con los archivos de las tres pantallas mezclados sería más
"puro" en la lectura, pero obliga a saltar entre carpetas para seguir el flujo de
una funcionalidad. Agrupando por pantalla, todo el código de los turnos está junto:
qué se pide, cómo se pinta y qué pasa al pulsar el botón. El flujo de datos de un
requisito se lee en tres archivos contiguos.

### Principales beneficios

- **Mantenibilidad:** la lógica de negocio queda aislada de la interfaz. Si cambia
  la regla de asignación de posiciones, se toca el modelo y no hay riesgo de romper
  el renderizado.
- **Facilidad para migrar o rediseñar:** si en algún momento se decide usar un
  framework, solo habría que rehacer las vistas; el resto de la lógica sigue
  intacta y es reutilizable.
- **Modularidad por dominio:** un cambio en la forma de la respuesta de la API
  obliga a tocar el modelo y nunca la vista.
- **Responsabilidades definidas:** cada componente evoluciona con menos dependencias
  de los demás, lo que facilita repartir el trabajo entre varias personas.

---

## 2. Sin frameworks ni proceso de compilación

El cliente usa **JavaScript ES6 nativo con módulos nativos** (`import`/`export`,
`type="module"`). No hay React, Vue ni Angular, ni siquiera jQuery, y tampoco hay
bundler, transpiler ni paso de build.

### Consecuencia práctica: hay que servir por HTTP

Los navegadores bloquean los módulos ES bajo el protocolo `file://`. Abrir
`index.html` con doble clic deja la página en blanco sin errores visibles en
muchos casos. Por eso el proyecto documenta el servidor estático de nginx
(Docker) o `python -m http.server` / `npx serve` como formas válidas de
servir `frontend/`.

### Principales beneficios

- **Cero dependencias que actualizar o auditar.** El código que se despliega es
  exactamente el que está escrito en el repositorio.
- **Sin configuración ni herramientas que aprender.** No hay `package.json`,
  lockfiles ni paso intermedio entre editar y recargar.
- **Carga inmediata.** No hay compilador que espere; útil en una sala de espera
  con conexión lenta.

---

## 3. Cero manejadores en línea

La guía del proyecto prohíbe explícitamente los atributos `onclick=""` y
`onsubmit=""` en el HTML. Todas las acciones se atan en JavaScript.

Para las tablas y listas que se repintan dinámicamente (los turnos del panel, el
tablero de llamados) se usa **delegación de eventos**: un único `addEventListener`
sobre el contenedor, y cada botón lleva un atributo `data-accion` que identifica
qué debe hacer.

```html
<button data-accion="llamar" data-id="42">Llamar</button>
```

```js
contenedor.addEventListener('click', (evento) => {
    const boton = evento.target.closest('[data-accion]');
    if (!boton) return;
    acciones[boton.dataset.accion](boton.dataset.id);
});
```

### Principales beneficios

- **Un solo manejador por pantalla** en vez de uno por botón, incluso cuando la
  lista tiene cientos de filas.
- **El HTML dinámico queda declarativo:** la vista describe *qué* hace cada
  botón, el controlador decide *cómo*.
- **Las plantillas pueden ser cadenas simples** sin tener que inyectar funciones,
  lo que a su vez evita `innerHTML` con lógica incrustada.

---

## 4. Escapado de todo dato externo

Todo dato que viene del backend pasa por la función `escapar()` de
`utils/dom.js` antes de insertarse como HTML.

```js
const ENTIDADES = { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' };

export const escapar = (valor) =>
    String(valor ?? '').replace(/[&<>"']/g, (caracter) => ENTIDADES[caracter]);
```

El `?? ''` garantiza que un `null` o `undefined` del backend se convierta en
cadena vacía en lugar de imprimir literalmente `undefined` en pantalla.

### Por qué no confiar en la base de datos

Los nombres y documentos de identidad los escribe una persona en un formulario
público. Un nombre como `<img src=x onerror=alert(1)>` es texto perfectamente
válido para la base de datos y una inyección para el tablero de llamados, que se
proyecta en una pantalla de la sala de espera. El escapado se aplica en el punto
de renderizado, no al guardar, porque ahí es donde ocurre el riesgo.

Como regla del proyecto, `innerHTML` solo aparece dentro de `views/` y de
`utils/dom.js`. Los modelos y los controladores nunca escriben en el DOM.

---

## 5. Configuración de la URL de la API

`js/config.js` resuelve a qué backend conectarse en este orden:

1. Si existe `localStorage['kora_api_url']`, se respeta.
2. Si la página está en `localhost` o `127.0.0.1`, se usa `http://localhost:8080`.
3. En cualquier otro dominio, se asume que la API se sirve en el mismo origen.

```js
const esLocal = ['localhost', '127.0.0.1', ''].includes(window.location.hostname);
return esLocal ? 'http://localhost:8080' : window.location.origin;
```

### Por qué ese orden

El caso 3 hace que el proyecto funcione **sin configuración** cuando se despliega
con un proxy inverso que sirve frontend y API en el mismo dominio. El caso 1
existe para desarrollo, cuando alguien apunta la interfaz a una API en otro
puerto o equipo. Y el caso 2 cubre el desarrollo local con el frontend servido
desde el 8081 y la API en el 8080.

### Lectura defensiva de `localStorage`

El acceso va envuelto en `try/catch`:

```js
const leerAjuste = (clave) => {
    try {
        return window.localStorage.getItem(clave);
    } catch {
        return null;
    }
};
```

Porque en **modo privado** de algunos navegadores y dentro de un `iframe` de
otro origen, `localStorage` lanza una excepción al leerlo. Sin este `try/catch`,
la aplicación entera fallaría al arrancar en vez de degradar a la autodetección.

---

## 6. Tiempo real con `EventSource` y respaldo automático

La conexión en vivo se resuelve en `models/tableroModel.js`. Se abre un
`EventSource` contra `/api/eventos`; si el flujo no se puede establecer, el modelo
cae a **polling** de `/api/tablero` cada 15 segundos y expone un estado
`degradado` que la vista usa para avisarle al usuario.

La degradación no es inmediata, porque un solo fallo suele ser una desconexión
momentánea y el propio `EventSource` ya reconecta solo:

```js
fuente.onerror = () => {
    fallos += 1;
    if (fuente) { fuente.close(); fuente = null; }

    // Tras varios intentos se confirma el respaldo por polling
    if (fallos >= 2) { iniciarPolling(); return; }

    // Un reintento corto antes de asumir que el flujo no va a volver
    temporizador = window.setTimeout(conectar, CONFIG.POLLING_INTERVAL / 3);
};
```

Y al revés: en cuanto el flujo se restablece, `onopen` cancela el intervalo de
polling para no tener dos mecanismos corriendo a la vez.

El modelo también contempla el caso de navegadores sin soporte de `EventSource`:
comprueba `typeof window.EventSource === 'undefined'` antes de construirlo y pasa
directo a polling.

### Principales beneficios

- **Código único para las tres pantallas.** El paciente, el tablero y el panel de
  personal consumen el mismo modelo; no hay una implementación por vista.
- **Degradación elegante:** si el flujo no está disponible, la aplicación sigue
  funcionando, más lenta, sin cambiar las vistas ni los controladores.
- **Sin librerías:** se respeta la restricción de no usar frameworks ni
  dependencias de cliente.

La clave de por qué funciona el respaldo es que **el servidor manda el tablero
completo en cada evento**. El cliente no reconstruye el estado ni decide qué
cambió: se limita a pintar lo que llega, y tanto el stream como el polling entregan
el mismo objeto.

---

## 7. Diseño visual: un design system con variables CSS

Todo el estilo vive en `css/styles.css`, construido sobre variables declaradas en
`:root`.

### Principales beneficios

- **Consistencia:** espaciado en múltiplos de 8, tipografía del sistema, bordes
  sobrios y sombras apenas perceptibles en toda la aplicación.
- **Un solo lugar que cambiar:** el color de acento o la escala de grises se
  ajustan en un punto, sin buscar literales por todo el proyecto.
- **Adaptación al entorno:** el tema oscuro se activa con `prefers-color-scheme`
  sin duplicar estilos, porque el tema alternativo solo redefine variables.
- **Accesibilidad:** estados `:focus-visible` visibles, contraste adecuado y
  respeto a `prefers-reduced-motion` para quien tenga activada esa preferencia
  en el sistema.
- **Responsive real:** solo Flexbox y CSS Grid, con puntos de corte en 960 px y
  768 px, para que la experiencia funcione igual en el celular de la sala de
  espera que en el escritorio de recepción.

---

## 8. Aviso sonoro con la Web Audio API

El aviso de llamado se genera con la **Web Audio API**, sin archivos de audio y
sin librerías.

```js
const oscillator = contexto.createOscillator();
oscillator.type = 'sine';
oscillator.frequency.value = 880;
```

### Por qué se genera y no se reproduce un archivo

Un `.mp3` o `.wav` obligaría a descargar un recurso binario extra y a esperar a
que cargue antes de poder avisar. Generando el tono con un oscilador, el sonido
está disponible en cuanto la Web Audio API está lista.

### Por qué es opcional y arranca apagado

Los navegadores **bloquean el audio automático** hasta que el usuario interactúa
con la página. Por eso el tablero tiene un interruptor de sonido que el usuario
activa conscientemente: intentar reproducirlo solo sería una promesa que el
navegador no puede cumplir.

---

## 9. Punto de entrada por página

Cada página HTML carga un único módulo con `type="module"`:

| Página | Entry point |
|---|---|
| `index.html` | `js/main.js` |
| `tablero.html` | `js/tablero.js` |
| `admin.html` | `js/admin.js` |

Cada entry point instancia su propio controlador. No existe un punto de entrada
único que conditionally cargue pantallas, ni variables globales como
`window.adminController`: el módulo ES ya encapsula su alcance, y una referencia
global solo serviría para que un `<script>` de consola manipulase la aplicación.

---

## 10. Limitaciones conocidas

Se documentan de forma explícita para que no se confundan con decisiones
consientes:

- **El token del panel se guarda en `localStorage`.** Es accesible desde cualquier
  script de la página, por lo que una vulnerabilidad de XSS implicaría robo de
  sesión. La mitigación correcta es una cookie `HttpOnly` con `SameSite`, algo que
  el backend no expone todavía. Como defensa intermedia, el proyecto escapa todo
  lo que se inserta en el DOM (§4).
- **`localStorage` guarda también la URL de la API**, no datos sensibles.
- **Sin caché de peticiones** más allá del catálogo de servicios, que se retiene
  5 minutos en memoria para no golpear el backend en cada cambio de pestaña.