-- rule: W01
-- reads: process_steps, failure_modes
-- message: Rework step {{.opNo}} {{.stepName}} has no PFMEA rows (IATF 16949 8.7.1.4).
-- returns: object_type, object_id, field, key, params (lihat docs/06-rules.md)
SELECT 'process_steps'::text, s.id, 'name'::text, ''::text,
       jsonb_build_object('opNo', s.op_no, 'stepName', s.name)
FROM process_steps s
WHERE s.package_id = @package_id
  AND s.kind = 'rework'
  AND NOT EXISTS (SELECT 1 FROM failure_modes fm WHERE fm.package_id = s.package_id AND fm.step_id = s.id)
