BEGIN;

-- Create role(s)

INSERT INTO public.roles(tenant_id, name)
    VALUES
        -- 1
        (1, 'org:admin'),
        -- 2
        (1, 'org:reader'),
        -- 3
        (1, 'org:team:admin');


INSERT INTO public.userroles(tenant_id, uuid, role_id)
    VALUES
        (1, 'e3cb82c9-6b37-4d13-8583-344e83ad74af', 1);

COMMIT;