-- rule: W02
-- reads: process_steps, step_flows
-- message: Inspection step {{.opNo}} {{.stepName}} has no NG branch with a disposition (rework, scrap or hold).
-- returns: object_type, object_id, field, key, params (lihat docs/06-rules.md)
SELECT 'process_steps'::text, s.id, 'ngFlow'::text, ''::text,
       jsonb_build_object('opNo', s.op_no, 'stepName', s.name)
FROM process_steps s
WHERE s.package_id = @package_id
  AND s.symbol IN ('inspection', 'operation_inspection')
  AND NOT EXISTS (SELECT 1 FROM step_flows f
                  WHERE f.package_id = s.package_id AND f.from_step_id = s.id
                    AND f.kind IN ('ng', 'rework', 'scrap')
                    AND coalesce(btrim(f.disposition), '') <> '')
