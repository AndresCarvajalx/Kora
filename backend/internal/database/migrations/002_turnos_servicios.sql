-- ============================================================
-- KORA · Migración 002 · Turnos por servicio
-- Agrega servicio, tipo de atención, documento, numeración de cola y
-- marcas de tiempo de llamada/atención/cancelación.
-- Es idempotente: se aplica también sobre bases creadas con la 001.
-- ============================================================

ALTER TABLE turnos ADD COLUMN IF NOT EXISTS servicio_id INTEGER REFERENCES servicios(id) ON DELETE SET NULL;
ALTER TABLE turnos ADD COLUMN IF NOT EXISTS tipo_atencion_id INTEGER REFERENCES tipos_atencion(id) ON DELETE SET NULL;
ALTER TABLE turnos ADD COLUMN IF NOT EXISTS cliente_documento VARCHAR(50);

-- Número visible del turno: "<CODIGO_SERVICIO>-<consecutivo>", ej. CAR-014
ALTER TABLE turnos ADD COLUMN IF NOT EXISTS numero_turno VARCHAR(12);
-- Posición inmutable dentro de la cola del día (servicio + tipo de atención)
ALTER TABLE turnos ADD COLUMN IF NOT EXISTS orden_cola INTEGER;
ALTER TABLE turnos ADD COLUMN IF NOT EXISTS cola_fecha DATE;

ALTER TABLE turnos ADD COLUMN IF NOT EXISTS llamado_en TIMESTAMP;
ALTER TABLE turnos ADD COLUMN IF NOT EXISTS atendido_en TIMESTAMP;
ALTER TABLE turnos ADD COLUMN IF NOT EXISTS cancelado_en TIMESTAMP;
ALTER TABLE turnos ADD COLUMN IF NOT EXISTS motivo_cancelacion VARCHAR(255);

-- Turnos antiguos (sin servicio) se les da una posición coherente por contacto
-- para que el orden de la cola nunca sea NULL dentro de una cola nueva.
UPDATE turnos SET numero_turno = 'LEG-' || LPAD(id::TEXT, 3, '0') WHERE numero_turno IS NULL;

-- El índice anterior era global por contacto; ahora es por servicio:
-- una persona puede tener un turno activo en cada servicio.
DROP INDEX IF EXISTS idx_turno_activo_por_contacto;

CREATE UNIQUE INDEX IF NOT EXISTS idx_turno_activo_contacto_servicio
    ON turnos(cliente_contacto, servicio_id)
    WHERE estado IN ('PENDIENTE', 'EN_ATENCION');

CREATE INDEX IF NOT EXISTS idx_turnos_estado ON turnos(estado);
CREATE INDEX IF NOT EXISTS idx_turnos_codigo ON turnos(numero_turno);
CREATE INDEX IF NOT EXISTS idx_turnos_contacto ON turnos(cliente_contacto);
CREATE INDEX IF NOT EXISTS idx_turnos_cola
    ON turnos(servicio_id, tipo_atencion_id, cola_fecha, estado, orden_cola);
CREATE INDEX IF NOT EXISTS idx_turnos_llamado ON turnos(llamado_en DESC NULLS LAST);