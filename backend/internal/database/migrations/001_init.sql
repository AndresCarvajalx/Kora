-- ============================================================
-- KORA · Migración 001 · Esquema inicial
-- Usuarios, catálogo de servicios, tipos de atención, turnos y
-- secuencias diarias de numeración por servicio.
-- ============================================================

CREATE TABLE IF NOT EXISTS usuarios (
    id SERIAL PRIMARY KEY,
    nombre VARCHAR(100) NOT NULL,
    email VARCHAR(150) UNIQUE NOT NULL,
    password_hash VARCHAR(255),
    rol VARCHAR(10) NOT NULL CHECK (rol IN ('CLIENTE', 'ADMIN')),
    creado_en TIMESTAMP DEFAULT NOW()
);

-- Catálogo de servicios médicos.
-- `codigo` es el prefijo que se muestra en el turno (ej. CAR-014).
CREATE TABLE IF NOT EXISTS servicios (
    id SERIAL PRIMARY KEY,
    codigo VARCHAR(4) NOT NULL UNIQUE,
    nombre VARCHAR(100) NOT NULL,
    descripcion VARCHAR(255),
    duracion_minutos INTEGER NOT NULL DEFAULT 15 CHECK (duracion_minutos > 0),
    activo BOOLEAN NOT NULL DEFAULT TRUE,
    orden INTEGER NOT NULL DEFAULT 0,
    creado_en TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Tipos de atención. Cada servicio define sus propios tipos.
CREATE TABLE IF NOT EXISTS tipos_atencion (
    id SERIAL PRIMARY KEY,
    servicio_id INTEGER NOT NULL REFERENCES servicios(id) ON DELETE CASCADE,
    nombre VARCHAR(60) NOT NULL,
    orden INTEGER NOT NULL DEFAULT 0,
    activo BOOLEAN NOT NULL DEFAULT TRUE,
    CONSTRAINT tipos_atencion_unico UNIQUE (servicio_id, nombre)
);

CREATE TABLE IF NOT EXISTS turnos (
    id SERIAL PRIMARY KEY,
    cliente_contacto VARCHAR(150) NOT NULL,
    cliente_nombre VARCHAR(100),
    fecha_hora TIMESTAMP NOT NULL,
    expira_en TIMESTAMP NOT NULL,
    estado VARCHAR(15) NOT NULL DEFAULT 'PENDIENTE'
        CHECK (estado IN ('PENDIENTE', 'EN_ATENCION', 'ATENDIDO', 'CANCELADO', 'VENCIDO')),
    creado_en TIMESTAMP DEFAULT NOW()
);

-- Consecutivo diario por servicio. La fila se incrementa con un
-- INSERT ... ON CONFLICT DO UPDATE (operación atómica) desde Go.
CREATE TABLE IF NOT EXISTS turno_secuencia (
    servicio_id INTEGER NOT NULL REFERENCES servicios(id) ON DELETE CASCADE,
    fecha DATE NOT NULL,
    ultimo INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (servicio_id, fecha)
);

-- Usuario admin por defecto (password: admin123)
INSERT INTO usuarios (nombre, email, password_hash, rol)
VALUES (
    'Admin Kora',
    'admin@kora.com',
    '$2a$10$Re5dJBVSsi.H0FNQiVsRS.rDEUrecU4v7AKEgHGC/QcEyNqGPPS2a',
    'ADMIN'
) ON CONFLICT (email) DO NOTHING;