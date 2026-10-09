-- rule: R04
-- reads: failure_chains, controls, control_library
-- message: D = {{.d}} (step {{.opNo}}) is too low for method {{.methods}}; the control library requires D of at least {{.dMin}}.
-- returns: object_type, object_id, field, key, params (lihat docs/06-rules.md)
SELECT 'failure_chains'::text, ch.id, 'd'::text, ''::text,
       jsonb_build_object('d', ch.d, 'opNo', s.op_no, 'dMin', x.d_min, 'methods', x.methods)
FROM failure_chains ch
JOIN failure_modes fm ON fm.id = ch.failure_mode_id
JOIN process_steps s ON s.id = fm.step_id
JOIN LATERAL (
  SELECT min(lib.d_min) AS d_min, string_agg(DISTINCT lib.name, ', ') AS methods
  FROM controls k
  JOIN control_library lib ON lib.id = k.control_library_id
  WHERE k.failure_chain_id = ch.id AND k.kind = 'detection' AND lib.d_min IS NOT NULL
) x ON x.d_min IS NOT NULL
WHERE ch.package_id = @package_id
  AND ch.d IS NOT NULL
  AND ch.d < x.d_min
