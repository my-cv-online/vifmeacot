-- rule: R01
-- reads: controls, failure_chains, failure_modes, cp_lines
-- message: Detection control "{{.control}}" (step {{.opNo}}, {{.charNo}}) is not in the Control Plan on the same step and characteristic.
-- returns: object_type, object_id, field, key, params (lihat docs/06-rules.md)
SELECT 'controls'::text, k.id, 'text'::text, ''::text,
       jsonb_build_object('control', k.text, 'opNo', s.op_no, 'charNo', c.char_no)
FROM controls k
JOIN failure_chains ch ON ch.id = k.failure_chain_id
JOIN failure_modes fm ON fm.id = ch.failure_mode_id
JOIN process_steps s ON s.id = fm.step_id
JOIN characteristics c ON c.id = fm.characteristic_id
WHERE k.package_id = @package_id
  AND k.kind = 'detection'
  AND btrim(k.text) <> ''
  AND NOT EXISTS (
    SELECT 1
    FROM cp_lines l
    LEFT JOIN controls k2 ON k2.id = l.control_id
    WHERE l.package_id = k.package_id
      AND l.step_id = fm.step_id
      AND l.characteristic_id = fm.characteristic_id
      AND (l.control_id = k.id OR lower(btrim(k2.text)) = lower(btrim(k.text)))   -- kontrol dengan teks sama di chain lain dianggap tercakup
  )
