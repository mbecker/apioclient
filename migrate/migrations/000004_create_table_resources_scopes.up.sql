CREATE TABLE IF NOT EXISTS resourcesscopes(
   resource_id integer,
   scope_id integer,
   created_at timestamptz NOT NULL DEFAULT NOW(),
   updated_at timestamptz,
   PRIMARY KEY(resource_id, scope_id)
);