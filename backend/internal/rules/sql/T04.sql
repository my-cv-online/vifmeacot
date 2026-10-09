-- rule: T04
-- reads: all content tables
-- message: Row "{{.label}}" was detached from Template General without a reason.
-- returns: object_type, object_id, field, key, params (lihat docs/06-rules.md)
SELECT x.object_type, x.id, 'detachReason'::text, ''::text, jsonb_build_object('label', x.label)
FROM (
  SELECT 'process_steps'::text AS object_type, id, op_no || ' ' || name AS label, detach_reason FROM process_steps
  WHERE package_id = @package_id AND sync_status = 'detached'
  UNION ALL
  SELECT 'step_flows'::text AS object_type, id, coalesce(label, kind::text) AS label, detach_reason FROM step_flows
  WHERE package_id = @package_id AND sync_status = 'detached'
  UNION ALL
  SELECT 'characteristics'::text AS object_type, id, char_no || ' ' || name AS label, detach_reason FROM characteristics
  WHERE package_id = @package_id AND sync_status = 'detached'
  UNION ALL
  SELECT 'failure_modes'::text AS object_type, id, text AS label, detach_reason FROM failure_modes
  WHERE package_id = @package_id AND sync_status = 'detached'
  UNION ALL
  SELECT 'failure_effects'::text AS object_type, id, text AS label, detach_reason FROM failure_effects
  WHERE package_id = @package_id AND sync_status = 'detached'
  UNION ALL
  SELECT 'failure_causes'::text AS object_type, id, text AS label, detach_reason FROM failure_causes
  WHERE package_id = @package_id AND sync_status = 'detached'
  UNION ALL
  SELECT 'failure_chains'::text AS object_type, id, id::text AS label, detach_reason FROM failure_chains
  WHERE package_id = @package_id AND sync_status = 'detached'
  UNION ALL
  SELECT 'controls'::text AS object_type, id, text AS label, detach_reason FROM controls
  WHERE package_id = @package_id AND sync_status = 'detached'
  UNION ALL
  SELECT 'cp_lines'::text AS object_type, id, coalesce(control_method, id::text) AS label, detach_reason FROM cp_lines
  WHERE package_id = @package_id AND sync_status = 'detached'
  UNION ALL
  SELECT 'reaction_plans'::text AS object_type, id, text AS label, detach_reason FROM reaction_plans
  WHERE package_id = @package_id AND sync_status = 'detached'
) x
WHERE coalesce(btrim(x.detach_reason), '') = ''
