CREATE TABLE IF NOT EXISTS roles(
   tenant_id integer REFERENCES tenants (tenant_id),
   role_id integer GENERATED ALWAYS AS IDENTITY (START WITH 1 INCREMENT BY 1) PRIMARY KEY,
   name VARCHAR(264) NOT NULL,
   created_at timestamptz NOT NULL DEFAULT NOW(),
   updated_at timestamptz,
   UNIQUE(tenant_id, name)
);