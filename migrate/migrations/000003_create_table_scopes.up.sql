CREATE TABLE IF NOT EXISTS scopes(
   scope_id integer GENERATED ALWAYS AS IDENTITY (START WITH 1 INCREMENT BY 1) PRIMARY KEY,
   tenant_id integer REFERENCES tenants (tenant_id),
   name VARCHAR (264) NOT NULL,
   defaultscope BOOLEAN NOT NULL DEFAULT FALSE,
   created_at timestamptz NOT NULL DEFAULT NOW(),
   updated_at timestamptz,
   UNIQUE(name, tenant_id)
);