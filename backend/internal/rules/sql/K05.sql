-- rule: K05
-- reads: cp_lines, process_steps
-- message: Machine/tool {{.unknown}} in the Control Plan is not listed on step {{.opNo}} in the PFD.
-- returns: object_type, object_id, field, key, params (lihat docs/06-rules.md)
SELECT 'cp_lines'::text, l.id, 'machines'::text, ''::text,
       jsonb_build_object('opNo', s.op_no,
                          'unknown', array_to_string(ARRAY(SELECT m FROM unnest(l.machines) AS m
                                                           WHERE NOT (m = ANY (s.machines))), ', '))
FROM cp_lines l
JOIN process_steps s ON s.id = l.step_id
WHERE l.package_id = @package_id
  AND NOT (l.machines <@ s.machines)
