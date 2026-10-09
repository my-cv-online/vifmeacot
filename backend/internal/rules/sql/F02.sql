-- rule: F02
-- reads: failure_modes
-- message: One cell contains more than one failure mode: "{{.text}}". Split it into separate rows.
-- returns: object_type, object_id, field, key, params (lihat docs/06-rules.md)
SELECT 'failure_modes'::text, fm.id, 'text'::text, ''::text,
       jsonb_build_object('text', fm.text, 'opNo', s.op_no)
FROM failure_modes fm
JOIN process_steps s ON s.id = fm.step_id
WHERE fm.package_id = @package_id
  -- kata sambung Inggris dan Indonesia, karena isi FMEA bisa ditulis dalam kedua bahasa
  AND fm.text ~* '(\s/\s|;|\n|\s&\s|\s(and|or|dan|atau)\s)'
