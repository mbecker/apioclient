CREATE RECURSIVE view resources_permissions (
		resource_id,
		name,
		type,
		parent, depth, path, is_leaf) as (
	SELECT
		resource_id,
		name,
		type,
		parent,
		0 AS depth,
		name::TEXT path,
		NOT EXISTS ( SELECT NULL
                             FROM resources gr
                             WHERE r.resource_id = gr.parent ) is_leaf
	FROM
		public.resources r 
	union all
		SELECT
			e.resource_id,
			e.name,
			e.type,
			e.parent,
			s.depth+ 1,
			s.path || ' > ' || e.name,
			NOT EXISTS ( SELECT NULL
                             FROM resources gr
                             WHERE s.resource_id = gr.parent )
		FROM
			public.resources e
		INNER JOIN resources_permissions s ON s.resource_id = e.parent
)