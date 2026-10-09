-- rule: T01
-- reads: process_steps, package_links, package_revisions
-- message: Mandatory general process {{.opNo}} {{.stepName}} (Template General rev {{.rev}}) is missing in the package.
-- returns: object_type, object_id, field, key, params (lihat docs/06-rules.md)
SELECT 'packages'::text, pk.id, 'steps'::text, g.id::text,
       jsonb_build_object('opNo', g.op_no, 'stepName', g.name, 'rev', lk.synced_rev)
FROM packages pk
JOIN package_links lk ON lk.model_package_id = pk.id
JOIN package_revisions r ON r.package_id = lk.general_package_id AND r.rev_no = lk.synced_rev
CROSS JOIN LATERAL jsonb_to_recordset(r.snapshot -> 'tables' -> 'process_steps')
  AS g(id uuid, op_no text, name text, general_mode text)
WHERE pk.id = @package_id
  AND g.general_mode = 'mandatory'
  AND NOT EXISTS (SELECT 1 FROM process_steps m WHERE m.package_id = pk.id AND m.source_id = g.id)
