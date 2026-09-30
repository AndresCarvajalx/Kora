# Decisiones Del Proyecto
A continuacion encontrara las decisiones tomadas por el equipo para desarrollar el proyecto y el por que de estas decisiones.

##  Arquitectura
Se eligio la arquitectura Modelo Vista Controlador (MVC).

Donde:
- **Modelo:** Maneja los datos, la lógica de negocio y el estado del sistema. No sabe cómo se ven las cosas en la interfaz de usuario

- **Vista:** Se encarga de la interfaz de usuario (HTML/CSS y manipulación del DOM). Muestra datos al usuario y captura eventos.

- **Controlador:** Escucha los eventos de la Vista, invoca la lógica del Modelo y le ordena a la Vista actualizarse

### Principales Beneficios
- **Mantenibilidad** ya que alisla la logica critica de la interfaz de usuario, toda la logica del modelo queda separada, por ejemplo la logica del agendamiento, de horarios, regla de notificaciones y etc. si toca cambiar alguna regla de notificaciones solo toca modificar el Modelo hay riesgo de rommper la interfaz

- **Facilidad para migrar/rediseñar** si en algun momento se decide cambiar el diseño de la interfaz o usar algun framework, solo tocaria rehacer la Vista, toda la logica de negocio seguiria intacta y reutilizable

- **Independencia:** Cada componente tiene responsabilidades definidas y puede evolucionar con menos dependencias de los demás.
- **Organización del trabajo:** La separación en componentes facilita distribuir tareas y mantener una estructura clara durante el desarrollo.