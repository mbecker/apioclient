CREATE TABLE IF NOT EXISTS rolepermissions(
   role_id integer,
   name VARCHAR(264) NOT NULL,
   ext_id VARCHAR(264) NOT NULL,
   created_at timestamptz NOT NULL DEFAULT NOW(),
   updated_at timestamptz,
   PRIMARY KEY(role_id)
);