BEGIN;
-- Resources
-- resource_id: 1
INSERT INTO public.resources(name, type) VALUES ('ruv', 'org');
-- resource_id: 2
INSERT INTO public.resources(name, type, parent) VALUES  ('kompass', 'team', 1);
-- resource_id: 3
INSERT INTO public.resources(name, type, parent) VALUES  ('racoon', 'team', 1);
-- resource_id: 4
INSERT INTO public.resources(name, type) VALUES ('huk', 'org');
-- resource_id: 5
INSERT INTO public.resources(name, type, parent) VALUES  ('auto', 'team', 4);
-- resource_id: 6
INSERT INTO public.resources(name, type, parent) VALUES  ('autoservice', 'service', 5);
-- resource_id: 7
INSERT INTO public.resources(name, type, parent) VALUES  ('fahrrad', 'team', 4);
-- resource_id: 8
INSERT INTO public.resources(name, type) VALUES  ('aok', 'org');
-- resource_id: 9
INSERT INTO public.resources(name, type, parent) VALUES  ('hartzer', 'team', 8);

-- SCOPES
-- scope_id: 1
INSERT INTO public.scopes(name) VALUES ('default');
-- scope_id: 2
INSERT INTO public.scopes(name) VALUES ('org:update');
-- scope_id: 3
INSERT INTO public.scopes(name) VALUES ('api:create');
-- scope_id: 4
INSERT INTO public.scopes(name) VALUES ('api:delete');
-- scope_id: 5
INSERT INTO public.scopes(name) VALUES ('api:update');
-- scope_id: 6
INSERT INTO public.scopes(name) VALUES ('api:read');
-- scope_id: 7
INSERT INTO public.scopes(name) VALUES ('member:update');
-- scope_id: 8
INSERT INTO public.scopes(name) VALUES ('member:invitation');
-- scope_id: 9
INSERT INTO public.scopes(name) VALUES ('service:update');

-- resourcesscopes
-- resourcesscope_id: 1 -- ruv->(defaultscope)
INSERT INTO public.resourcesscopes(
	resource_id, scope_id, defaultscope)
	VALUES (1, 1, TRUE);
-- resourcesscope_id: 2 -- ruv->(defaultscope)
INSERT INTO public.resourcesscopes(
	resource_id, scope_id, defaultscope)
	VALUES (1, 2, FALSE);
-- resourcesscope_id: 3 -- ruv->kompass->(defaultscope)
INSERT INTO public.resourcesscopes(
	resource_id, scope_id, defaultscope)
	VALUES (2, 1, TRUE);
-- resourcesscope_id: 4 -- ruv->kompass->(api:create)
INSERT INTO public.resourcesscopes(
	resource_id, scope_id, defaultscope)
	VALUES (2, 3, FALSE);
-- resourcesscope_id: 5 -- ruv->kompass->(api:delete)
INSERT INTO public.resourcesscopes(
	resource_id, scope_id, defaultscope)
	VALUES (2, 4, FALSE);
-- resourcesscope_id: 6 -- ruv->kompass->(api:update)
INSERT INTO public.resourcesscopes(
	resource_id, scope_id, defaultscope)
	VALUES (2, 5, FALSE);
-- resourcesscope_id: 7 -- kompass->api:read
INSERT INTO public.resourcesscopes(
	resource_id, scope_id, defaultscope)
	VALUES (2, 6, FALSE);
-- resourcesscope_id: 8 -- ruv->racoon->api:read
INSERT INTO public.resourcesscopes(
	resource_id, scope_id, defaultscope)
	VALUES (3, 6, FALSE);
-- resourcesscope_id: 9 -- huk->default(defaultscope)
INSERT INTO public.resourcesscopes(
	resource_id, scope_id, defaultscope)
	VALUES (4, 1, TRUE);
-- resourcesscope_id: 10 -- huk->default(defaultscope)
INSERT INTO public.resourcesscopes(
	resource_id, scope_id, defaultscope)
	VALUES (4, 2, FALSE);
-- resourcesscope_id: 11 -- huk->(member:update)
INSERT INTO public.resourcesscopes(
	resource_id, scope_id, defaultscope)
	VALUES (4, 7, FALSE);
-- resourcesscope_id: 12 -- huk->(member:invitation)
INSERT INTO public.resourcesscopes(
	resource_id, scope_id, defaultscope)
	VALUES (4, 8, FALSE);
-- resourcesscope_id: 13 -- huk->car->default(defaultscope)
INSERT INTO public.resourcesscopes(
	resource_id, scope_id, defaultscope)
	VALUES (5, 1, TRUE);
-- resourcesscope_id: 14 -- huk->car->(api:create)
INSERT INTO public.resourcesscopes(
	resource_id, scope_id, defaultscope)
	VALUES (5, 3, FALSE);
-- resourcesscope_id: 15 -- huk->car->autoservice(service:updat)
INSERT INTO public.resourcesscopes(
	resource_id, scope_id, defaultscope)
	VALUES (6, 9, FALSE);
-- resourcesscope_id: 16 -- huk->fahrrad->(default)
INSERT INTO public.resourcesscopes(
	resource_id, scope_id, defaultscope)
	VALUES (7, 1, FALSE);
-- resourcesscope_id: 16 -- aok->(default)
INSERT INTO public.resourcesscopes(
	resource_id, scope_id, defaultscope)
	VALUES (8, 1, TRUE);
-- resourcesscope_id: 16 -- aok->(org:update)
INSERT INTO public.resourcesscopes(
	resource_id, scope_id, defaultscope)
	VALUES (8, 2, FALSE);
-- resourcesscope_id: 16 -- aok->hartzer(default)
INSERT INTO public.resourcesscopes(
	resource_id, scope_id, defaultscope)
	VALUES (9, 1, FALSE);
-- resourcesscope_id: 16 -- aok->hartzer(api:create)
INSERT INTO public.resourcesscopes(
	resource_id, scope_id, defaultscope)
	VALUES (9, 3, FALSE);

--  USER PERMISSIONS
-- USER->RUV->kompass(all permission created above)
INSERT INTO public.userpermissions(
	uuid, resourcesscope_id)
	VALUES ('e3cb82c9-6b37-4d13-8583-344e83ad74af', 4);
INSERT INTO public.userpermissions(
	uuid, resourcesscope_id)
	VALUES ('e3cb82c9-6b37-4d13-8583-344e83ad74af', 5);
INSERT INTO public.userpermissions(
	uuid, resourcesscope_id)
	VALUES ('e3cb82c9-6b37-4d13-8583-344e83ad74af', 6);
INSERT INTO public.userpermissions(
	uuid, resourcesscope_id)
	VALUES ('e3cb82c9-6b37-4d13-8583-344e83ad74af', 7);
INSERT INTO public.userpermissions(
	uuid, resourcesscope_id)
	VALUES ('e3cb82c9-6b37-4d13-8583-344e83ad74af', 8);
-- USER->huk(all permission created above)
INSERT INTO public.userpermissions(
	uuid, resourcesscope_id)
	VALUES ('e3cb82c9-6b37-4d13-8583-344e83ad74af', 11);
INSERT INTO public.userpermissions(
	uuid, resourcesscope_id)
	VALUES ('e3cb82c9-6b37-4d13-8583-344e83ad74af', 12);
-- USER->huk->car(api:create)
-- INSERT INTO public.userpermissions(
-- 	uuid, resourcesscope_id)
-- 	VALUES ('e3cb82c9-6b37-4d13-8583-344e83ad74af', 14);

-- USER->huk->car->servicecar(service:update)
INSERT INTO public.userpermissions(
	uuid, resourcesscope_id)
	VALUES ('e3cb82c9-6b37-4d13-8583-344e83ad74af', 15);


COMMIT;