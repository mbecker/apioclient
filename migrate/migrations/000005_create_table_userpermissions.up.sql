CREATE TABLE IF NOT EXISTS userpermissions(
   tenant_id integer REFERENCES tenants (tenant_id),
   uuid UUID,
   resourcesscope_id integer REFERENCES resourcesscopes (resourcesscope_id),
   created_at timestamptz NOT NULL DEFAULT NOW(),
   updated_at timestamptz,
   PRIMARY KEY(resourcesscope_id, uuid)
);