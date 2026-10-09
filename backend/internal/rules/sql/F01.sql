-- rule: F01
-- reads: failure_modes, failure_effects, failure_causes, failure_chains
-- message: {{.label}} is empty (step {{.opNo}}).
-- returns: object_type, object_id, field, key, params (lihat docs/06-rules.md)
SELECT 'failure_modes'::text, fm.id, 'text'::text, ''::text,
       jsonb_build_object('label', 'Failure mode', 'opNo', s.op_no)
FROM failure_modes fm JOIN process_steps s ON s.id = fm.step_id
WHERE fm.package_id = @package_id AND btrim(fm.text) = ''
UNION ALL
SELECT 'failure_modes'::text, fm.id, 'effects'::text, ''::text,
       jsonb_build_object('label', 'Failure effect', 'opNo', s.op_no)
FROM failure_modes fm JOIN process_steps s ON s.id = fm.step_id
WHERE fm.package_id = @package_id
  AND NOT EXISTS (SELECT 1 FROM failure_effects e WHERE e.failure_mode_id = fm.id AND btrim(e.text) <> '')
UNION ALL
SELECT 'failure_causes'::text, fc.id, 'text'::text, ''::text,
       jsonb_build_object('label', 'Failure cause', 'opNo', s.op_no)
FROM failure_causes fc
JOIN failure_modes fm ON fm.id = fc.failure_mode_id
JOIN process_steps s ON s.id = fm.step_id
WHERE fc.package_id = @package_id AND btrim(fc.text) = ''
UNION ALL
SELECT 'failure_chains'::text, ch.id, f.field, f.field,
       jsonb_build_object('label', f.label, 'opNo', s.op_no)
FROM failure_chains ch
JOIN failure_modes fm ON fm.id = ch.failure_mode_id
JOIN process_steps s ON s.id = fm.step_id
CROSS JOIN LATERAL (VALUES ('s', 'Severity (S)', ch.s IS NULL),
                           ('o', 'Occurrence (O)', ch.o IS NULL),
                           ('d', 'Detection (D)', ch.d IS NULL)) AS f(field, label, missing)
WHERE ch.package_id = @package_id AND f.missing
