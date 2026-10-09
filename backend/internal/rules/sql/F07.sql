-- rule: F07
-- reads: failure_chains, failure_effects, actions
-- message: S = {{.s}} (step {{.opNo}}) without a recommended action or justification.
-- returns: object_type, object_id, field, key, params (lihat docs/06-rules.md)
SELECT 'failure_chains'::text, ch.id, 's'::text, ''::text,
       jsonb_build_object('s', ch.s, 'opNo', s.op_no)
FROM failure_chains ch
JOIN failure_modes fm ON fm.id = ch.failure_mode_id
JOIN process_steps s ON s.id = fm.step_id
WHERE ch.package_id = @package_id
  AND ch.s >= 9
  AND coalesce(btrim(ch.justification), '') = ''
  AND NOT EXISTS (SELECT 1 FROM actions a
                  WHERE a.failure_chain_id = ch.id AND a.status <> 'cancelled' AND btrim(a.text) <> '')
