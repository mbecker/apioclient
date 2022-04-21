WITH RECURSIVE subresources AS (
	SELECT
		resource_id,
		name,
		type,
		parent
	FROM
		public.resources r 
	UNION
		SELECT
			e.resource_id,
			e.name,
			e.type,
			e.parent
		FROM
			public.resources e
		INNER JOIN subresources s ON s.parent = e.resource_id 
)
select u.uuid, s.resource_id, s.name, s.type, s.parent, s2."name", r.defaultscope FROM subresources s left join public.resourcesscopes r on s.resource_id = r.resource_id left join public.scopes s2 on r.scope_id = s2.scope_id left join public.userpermissions u on r.resourcesscope_id = u.resourcesscope_id
where (u.uuid = 'e3cb82c9-6b37-4d13-8583-344e83ad74af' or r.defaultscope = true);
--select * from subresources;
--SELECT STRING_AGG(name,':' ORDER BY level DESC) AS path FROM subresources;
