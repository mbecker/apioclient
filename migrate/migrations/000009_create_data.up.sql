BEGIN;
-- Resources
-- resource_id: 1
INSERT INTO public.resources(name, type)
	VALUES
		-- 1
		('ruv', 'org'),
		-- 2
		('ruv:kompass', 'org:team'),
		-- 3
		('ruv:racoon', 'org:team'),
		-- 4
		('huk', 'org'),
		-- 5
		('huk:car', 'org:team'),
		-- 6
		('huk:car:service1', 'org:team:service');

-- resource_id: 2
-- INSERT INTO public.resources(name, type) VALUES  ('ruv:kompass', 'org:team');
-- resource_id: 3
-- INSERT INTO public.resources(name, type) VALUES  ('ruv:racoon', 'org:team');

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


INSERT INTO public.scopes(name)
	VALUES 
		-- 7
		('org:read'),
		-- 8
		('team:create'),
		-- 9
		('service:read');


-- resourcesscopes
-- resourcesscope_id: 1 -- ruv->(defaultscope)
INSERT INTO public.resourcesscopes(
	resource_id, scope_id, defaultscope)
	VALUES (2, 2, FALSE);
INSERT INTO public.resourcesscopes(
	resource_id, scope_id, defaultscope)
	VALUES (2, 3, FALSE);
INSERT INTO public.resourcesscopes(
	resource_id, scope_id, defaultscope)
	VALUES (2, 4, FALSE);
INSERT INTO public.resourcesscopes(
	resource_id, scope_id, defaultscope)
	VALUES (2, 5, FALSE);
INSERT INTO public.resourcesscopes(
	resource_id, scope_id, defaultscope)
	VALUES (2, 6, FALSE);
INSERT INTO public.resourcesscopes(
	resource_id, scope_id, defaultscope)
	VALUES (3, 6, FALSE);

INSERT INTO public.resourcesscopes(
	resource_id, scope_id, defaultscope)
	VALUES 
	-- 7 huk->(org:update)
	(4, 2, FALSE),
	-- 8 huk->(org:read)
	(4, 7, FALSE),
	-- 9 huk->(team:create)
	(4, 8, FALSE),
	-- 10 huk->car(team:read)
	(5, 3, FALSE),
	-- 11 huk->car->service(service:read)
	(6, 9, FALSE);




--  USER PERMISSIONS
-- USER->RUV->kompass(all permission created above)
INSERT INTO public.userpermissions(
	uuid, resourcesscope_id)
	VALUES  ('e3cb82c9-6b37-4d13-8583-344e83ad74af', 1),
          ('e3cb82c9-6b37-4d13-8583-344e83ad74af', 2),
          ('e3cb82c9-6b37-4d13-8583-344e83ad74af', 3),
          ('e3cb82c9-6b37-4d13-8583-344e83ad74af', 4),
          ('e3cb82c9-6b37-4d13-8583-344e83ad74af', 5),
          ('e3cb82c9-6b37-4d13-8583-344e83ad74af', 6),
		  ('e3cb82c9-6b37-4d13-8583-344e83ad74af', 11),
		  ('e3cb82c9-6b37-4d13-8583-344e83ad74af', 10);

COMMIT;