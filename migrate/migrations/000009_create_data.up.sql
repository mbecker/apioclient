BEGIN;
-- Tneants
INSERT INTO public.tenants(name) VALUES ('ruv');
-- Resources
-- resource_id: 1
INSERT INTO public.resources(name, type, tenant_id)
	VALUES
		-- 1
		('ruv', 'org', 1),
		-- 2
		('ruv:kompass', 'org:team', 1),
		-- 3
		('ruv:racoon', 'org:team', 1),
		-- 4
		('huk', 'org', 1),
		-- 5
		('huk:car', 'org:team', 1),
		-- 6
		('huk:car:service1', 'org:team:service', 1);

-- resource_id: 2
-- INSERT INTO public.resources(name, type) VALUES  ('ruv:kompass', 'org:team');
-- resource_id: 3
-- INSERT INTO public.resources(name, type) VALUES  ('ruv:racoon', 'org:team');

-- SCOPES
-- scope_id: 1
INSERT INTO public.scopes(name, tenant_id) VALUES ('default', 1);
-- scope_id: 2
INSERT INTO public.scopes(name, tenant_id) VALUES ('org:update', 1);
-- scope_id: 3
INSERT INTO public.scopes(name, tenant_id) VALUES ('api:create', 1);
-- scope_id: 4
INSERT INTO public.scopes(name, tenant_id) VALUES ('api:delete', 1);
-- scope_id: 5
INSERT INTO public.scopes(name, tenant_id) VALUES ('api:update', 1);
-- scope_id: 6
INSERT INTO public.scopes(name, tenant_id) VALUES ('api:read', 1);


INSERT INTO public.scopes(name, tenant_id)
	VALUES 
		-- 7
		('org:read', 1),
		-- 8
		('team:create', 1),
		-- 9
		('service:read', 1);


-- resourcesscopes
-- resourcesscope_id: 1 -- ruv->(defaultscope)
INSERT INTO public.resourcesscopes(
	resource_id, scope_id, defaultscope, tenant_id)
	VALUES (2, 2, FALSE, 1);
INSERT INTO public.resourcesscopes(
	resource_id, scope_id, defaultscope, tenant_id)
	VALUES (2, 3, FALSE, 1);
INSERT INTO public.resourcesscopes(
	resource_id, scope_id, defaultscope, tenant_id)
	VALUES (2, 4, FALSE, 1);
INSERT INTO public.resourcesscopes(
	resource_id, scope_id, defaultscope, tenant_id)
	VALUES (2, 5, FALSE, 1);
INSERT INTO public.resourcesscopes(
	resource_id, scope_id, defaultscope, tenant_id)
	VALUES (2, 6, FALSE, 1);
INSERT INTO public.resourcesscopes(
	resource_id, scope_id, defaultscope, tenant_id)
	VALUES (3, 6, FALSE, 1);

INSERT INTO public.resourcesscopes(
	resource_id, scope_id, defaultscope, tenant_id)
	VALUES 
	-- 7 huk->(org:update)
	(4, 2, FALSE, 1),
	-- 8 huk->(org:read)
	(4, 7, FALSE, 1),
	-- 9 huk->(team:create)
	(4, 8, FALSE, 1),
	-- 10 huk->car(team:read)
	(5, 3, FALSE, 1),
	-- 11 huk->car->service(service:read)
	(6, 9, FALSE, 1);




--  USER PERMISSIONS
-- USER->RUV->kompass(all permission created above)
INSERT INTO public.userpermissions(
	uuid, resourcesscope_id, tenant_id)
	VALUES  ('e3cb82c9-6b37-4d13-8583-344e83ad74af', 1, 1),
          ('e3cb82c9-6b37-4d13-8583-344e83ad74af', 2, 1),
          ('e3cb82c9-6b37-4d13-8583-344e83ad74af', 3, 1),
          ('e3cb82c9-6b37-4d13-8583-344e83ad74af', 4, 1),
          ('e3cb82c9-6b37-4d13-8583-344e83ad74af', 5, 1),
          ('e3cb82c9-6b37-4d13-8583-344e83ad74af', 6, 1),
		  ('e3cb82c9-6b37-4d13-8583-344e83ad74af', 11, 1),
		  ('e3cb82c9-6b37-4d13-8583-344e83ad74af', 10, 1);

COMMIT;