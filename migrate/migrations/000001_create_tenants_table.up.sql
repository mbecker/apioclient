CREATE TABLE IF NOT EXISTS tenants(
   tenant_id integer GENERATED ALWAYS AS IDENTITY (START WITH 1000 INCREMENT BY 100) PRIMARY KEY,
   name VARCHAR (264) NOT NULL,
   created_at timestamptz NOT NULL DEFAULT NOW(),
   updated_at timestamptz,
   UNIQUE(name)
);