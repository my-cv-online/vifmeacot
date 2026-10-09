-- rule: R03
-- reads: cp_lines
-- message: CP line {{.charNo}} (step {{.opNo}}) is not linked to a failure mode or control in the PFMEA.
-- returns: object_type, object_id, field, key, params (lihat docs/06-rules.md)
SELECT 'cp_lines'::text, l.id, 'controlId'::text, ''::text,
       jsonb_build_object('charNo', c.char_no, 'opNo', s.op_no)
FROM cp_lines l
JOIN process_steps s ON s.id = l.step_id
JOIN characteristics c ON c.id = l.characteristic_id
WHERE l.package_id = @package_id
  AND l.control_id IS NULL
  AND l.failure_mode_id IS NULL
