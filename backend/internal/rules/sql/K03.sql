-- rule: K03
-- reads: failure_modes, cp_lines, characteristics
-- message: {{.what}} in step {{.opNo}} references characteristic {{.charNo}} of step {{.charOpNo}}.
-- returns: object_type, object_id, field, key, params (lihat docs/06-rules.md)
SELECT 'failure_modes'::text, fm.id, 'characteristicId'::text, ''::text,
       jsonb_build_object('what', 'Failure mode', 'opNo', s.op_no, 'charNo', c.char_no, 'charOpNo', cs.op_no)
FROM failure_modes fm
JOIN process_steps s ON s.id = fm.step_id
JOIN characteristics c ON c.id = fm.characteristic_id
JOIN process_steps cs ON cs.id = c.step_id
WHERE fm.package_id = @package_id AND c.step_id <> fm.step_id
UNION ALL
SELECT 'cp_lines'::text, l.id, 'characteristicId'::text, ''::text,
       jsonb_build_object('what', 'CP line', 'opNo', s.op_no, 'charNo', c.char_no, 'charOpNo', cs.op_no)
FROM cp_lines l
JOIN process_steps s ON s.id = l.step_id
JOIN characteristics c ON c.id = l.characteristic_id
JOIN process_steps cs ON cs.id = c.step_id
WHERE l.package_id = @package_id AND c.step_id <> l.step_id
