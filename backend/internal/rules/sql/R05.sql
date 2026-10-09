-- rule: R05
-- reads: failure_chains, controls
-- message: O = {{.o}} (step {{.opNo}}) without a prevention control or history data.
-- returns: object_type, object_id, field, key, params (lihat docs/06-rules.md)
SELECT 'failure_chains'::text, ch.id, 'o'::text, ''::text,
       jsonb_build_object('o', ch.o, 'opNo', s.op_no)
FROM failure_chains ch
JOIN failure_modes fm ON fm.id = ch.failure_mode_id
JOIN process_steps s ON s.id = fm.step_id
WHERE ch.package_id = @package_id
  AND ch.o IS NOT NULL AND ch.o <= 3
  AND coalesce(btrim(ch.o_evidence), '') = ''
  AND NOT EXISTS (SELECT 1 FROM controls k
                  WHERE k.failure_chain_id = ch.id AND k.kind = 'prevention' AND btrim(k.text) <> '')
