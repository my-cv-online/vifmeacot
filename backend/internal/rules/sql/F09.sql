-- rule: F09
-- reads: actions
-- message: Action "{{.action}}" is done but {{.label}} is not filled in.
-- returns: object_type, object_id, field, key, params (lihat docs/06-rules.md)
SELECT 'actions'::text, a.id, f.field, f.field,
       jsonb_build_object('action', a.text, 'label', f.label)
FROM actions a
CROSS JOIN LATERAL (VALUES ('newS', 'new S', a.new_s IS NULL),
                           ('newO', 'new O', a.new_o IS NULL),
                           ('newD', 'new D', a.new_d IS NULL)) AS f(field, label, missing)
WHERE a.package_id = @package_id
  AND a.status = 'done'
  AND f.missing
