-- rule: C01
-- reads: cp_lines, reaction_plans, characteristics
-- message: {{.label}} is empty on CP line {{.charNo}} (step {{.opNo}}).
-- returns: object_type, object_id, field, key, params (lihat docs/06-rules.md)
SELECT 'cp_lines'::text, l.id, f.field, f.field,
       jsonb_build_object('label', f.label, 'charNo', c.char_no, 'opNo', s.op_no)
FROM cp_lines l
JOIN process_steps s ON s.id = l.step_id
JOIN characteristics c ON c.id = l.characteristic_id
LEFT JOIN reaction_plans rp ON rp.cp_line_id = l.id
CROSS JOIN LATERAL (VALUES
  ('spec',          'Specification/Tolerance',           coalesce(btrim(c.spec), '') = ''),
  ('evalTechnique', 'Evaluation/Measurement Technique',  coalesce(btrim(l.eval_technique), '') = ''),
  ('sampleSize',    'Sample Size',                       coalesce(btrim(l.sample_size), '') = ''),
  ('sampleFreq',    'Sample Frequency',                  coalesce(btrim(l.sample_freq), '') = ''),
  ('controlMethod', 'Control Method',                    coalesce(btrim(l.control_method), '') = ''),
  ('reactionPlan',  'Reaction Plan',                     coalesce(btrim(rp.text), '') = '' AND coalesce(btrim(rp.isolation), '') = '')
) AS f(field, label, missing)
WHERE l.package_id = @package_id
  AND f.missing
