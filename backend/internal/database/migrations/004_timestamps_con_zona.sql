-- Los instantes de los turnos pasan a TIMESTAMPTZ.
-- Con TIMESTAMP (sin zona), pgx devuelve la hora local como si fuera UTC y, en
-- servidores con zona horaria distinta de UTC (p. ej. America/Bogota), los
-- turnos aparecían vencidos apenas se creaban.
ALTER TABLE turnos
    ALTER COLUMN fecha_hora   TYPE TIMESTAMPTZ,
    ALTER COLUMN expira_en    TYPE TIMESTAMPTZ,
    ALTER COLUMN creado_en    TYPE TIMESTAMPTZ,
    ALTER COLUMN llamado_en   TYPE TIMESTAMPTZ,
    ALTER COLUMN atendido_en  TYPE TIMESTAMPTZ,
    ALTER COLUMN cancelado_en TYPE TIMESTAMPTZ;
