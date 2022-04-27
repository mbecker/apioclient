CREATE TABLE IF NOT EXISTS resources(
   resource_id integer GENERATED ALWAYS AS IDENTITY (START WITH 1 INCREMENT BY 1) PRIMARY KEY,
   tenant_id integer REFERENCES tenants (tenant_id),
   name VARCHAR (264) NOT NULL,
   type VARCHAR (264),
   -- parent integer,
   created_at timestamptz NOT NULL DEFAULT NOW(),
   updated_at timestamptz,
   UNIQUE(name, tenant_id)
   -- CONSTRAINT fk_resources_parent FOREIGN KEY(parent) REFERENCES resources(resource_id)
);