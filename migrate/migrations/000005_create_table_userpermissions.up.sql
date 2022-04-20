CREATE TABLE IF NOT EXISTS userpermissions(
   resourcescope_id integer,
   created_at timestamptz NOT NULL DEFAULT NOW(),
   updated_at timestamptz,
   uuid UUID,PRIMARY KEY(resourcescope_id, uuid)
);