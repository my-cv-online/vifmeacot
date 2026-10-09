-- rule: R02
-- reads: controls, cp_lines
-- message: Prevention control "{{.control}}" (step {{.opNo}}) is not in the Control Plan and not marked as a system control.
-- returns: object_type, object_id, field, key, params (lihat docs/06-rules.md)
SELECT 'controls'::text, k.id, 'text'::text, ''::text,
       jsonb_build_object('control', k.text, 'opNo', s.op_no)
FROM controls k
JOIN failure_chains ch ON ch.id = k.failure_chain_id
JOIN failure_modes fm ON fm.id = ch.failure_mode_id
JOIN process_steps s ON s.id = fm.step_id
WHERE k.package_id = @package_id
  AND k.kind = 'prevention'
  AND btrim(k.text) <> ''
  AND NOT k.is_system_control
  AND NOT EXISTS (
    SELECT 1
    FROM cp_lines l
    LEFT JOIN controls k2 ON k2.id = l.control_id
    WHERE l.package_id = k.package_id
      AND (l.control_id = k.id OR (l.step_id = fm.step_id AND lower(btrim(k2.text)) = lower(btrim(k.text))))
  )
