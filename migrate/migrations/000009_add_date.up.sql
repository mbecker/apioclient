BEGIN;
-- Resources
-- resource_id: 1
INSERT INTO public.resources(name, type) VALUES ('ruv', 'org');
-- resource_id: 2
INSERT INTO public.resources(name, type, parent) VALUES  ('kompass', 'team', 1);

-- SCOPES
-- scope_id: 1
INSERT INTO public.scopes(name) VALUES ('default');
-- scope_id: 2
INSERT INTO public.scopes(name) VALUES ('org:update');
-- scope_id: 3
INSERT INTO public.scopes(name) VALUES ('member:create');

-- resourcesscopes
-- resourcesscope_id: 1 -- ruv->default(defaultscope)
INSERT INTO public.resourcesscopes(
	resource_id, scope_id, defaultscope)
	VALUES (1, 1, TRUE);
-- resourcesscope_id: 2 -- ruv->default(defaultscope)
INSERT INTO public.resourcesscopes(
	resource_id, scope_id, defaultscope)
	VALUES (1, 2, FALSE);
-- resourcesscope_id: 3 -- kompass->default(defaultscope)
INSERT INTO public.resourcesscopes(
	resource_id, scope_id, defaultscope)
	VALUES (2, 1, TRUE);
-- resourcesscope_id: 4 -- kompass->member:create
INSERT INTO public.resourcesscopes(
	resource_id, scope_id, defaultscope)
	VALUES (2, 3, FALSE);

--  USER PERMISSIONS
INSERT INTO public.userpermissions(
	uuid, resourcesscope_id)
	VALUES ('e3cb82c9-6b37-4d13-8583-344e83ad74af', 4);


COMMIT;