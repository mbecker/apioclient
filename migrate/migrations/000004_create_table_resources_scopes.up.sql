CREATE TABLE IF NOT EXISTS resourcesscopes(
   tenant_id integer REFERENCES tenants (tenant_id),
   resourcesscope_id integer GENERATED ALWAYS AS IDENTITY (START WITH 1 INCREMENT BY 1),
   resource_id integer REFERENCES resources (resource_id),
   scope_id integer REFERENCES scopes (scope_id),
   created_at timestamptz NOT NULL DEFAULT NOW(),
   updated_at timestamptz,
   PRIMARY KEY(resourcesscope_id),
   UNIQUE(resource_id, scope_id, tenant_id)
);