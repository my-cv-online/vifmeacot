-- rule: F08
-- reads: failure_chains, failure_effects, actions, customers
-- message: RPN {{.rpn}} (step {{.opNo}}) exceeds the customer threshold {{.threshold}} without action or justification.
-- returns: object_type, object_id, field, key, params (lihat docs/06-rules.md)
SELECT 'failure_chains'::text, ch.id, 'rpn'::text, ''::text,
       jsonb_build_object('rpn', ch.rpn, 'threshold', cu.rpn_action_threshold, 'opNo', s.op_no)
FROM failure_chains ch
JOIN packages pk ON pk.id = ch.package_id
JOIN customers cu ON cu.id = pk.customer_id AND cu.rpn_action_threshold IS NOT NULL
JOIN failure_modes fm ON fm.id = ch.failure_mode_id
JOIN process_steps s ON s.id = fm.step_id
WHERE ch.package_id = @package_id
  AND ch.rpn >= cu.rpn_action_threshold
  AND coalesce(btrim(ch.justification), '') = ''
  AND NOT EXISTS (SELECT 1 FROM actions a
                  WHERE a.failure_chain_id = ch.id AND a.status <> 'cancelled' AND btrim(a.text) <> '')
