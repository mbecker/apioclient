CREATE TABLE IF NOT EXISTS userroles(
   tenant_id integer REFERENCES tenants (tenant_id),
   uuid UUID,
   role_id integer REFERENCES roles (role_id),
   created_at timestamptz NOT NULL DEFAULT NOW(),
   updated_at timestamptz,
   PRIMARY KEY(role_id, uuid)
);