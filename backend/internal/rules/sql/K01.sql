-- rule: K01
-- reads: process_steps, failure_modes
-- message: Step {{.opNo}} {{.stepName}} is not analysed in the PFMEA.
-- returns: object_type, object_id, field, key, params (lihat docs/06-rules.md)
SELECT 'process_steps'::text AS object_type, s.id AS object_id, 'name'::text AS field, ''::text AS key,
       jsonb_build_object('opNo', s.op_no, 'stepName', s.name) AS params
FROM process_steps s
WHERE s.package_id = @package_id
  AND s.kind <> 'rework'                       -- step rework dicek oleh W01
  AND NOT s.not_analyzed
  AND NOT EXISTS (SELECT 1 FROM failure_modes fm WHERE fm.package_id = s.package_id AND fm.step_id = s.id)
