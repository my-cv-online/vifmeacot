-- Implementasi acuan read model PFMEA (PfmeaView di api/openapi.yaml).
-- Satu statement, satu round trip; Go langsung mengalirkan byte hasilnya tanpa men-decode.
-- Parameter: $1 = id paket. Jalankan di dalam transaksi REPEATABLE READ READ ONLY.
-- Sudah divalidasi terhadap db/seed/demo.sql dan paket buatan berisi 3.000 chain saat
-- spesifikasi ditulis (docs/03-architecture.md §4, §9). Tulis view PFD dan Control Plan dengan cara yang sama.
WITH pkg AS (
  SELECT p.*, c.code AS customer_code, c.rpn_action_threshold,
         coalesce(lk.general_package_id, CASE WHEN p.kind = 'general' THEN p.id END) AS policy_package_id
  FROM packages p
  LEFT JOIN customers c ON c.id = p.customer_id
  LEFT JOIN package_links lk ON lk.model_package_id = p.id
  WHERE p.id = $1
)
SELECT json_build_object(
  'package', (SELECT json_build_object('id', id, 'kind', kind, 'code', code, 'name', name,
                                       'customerId', customer_id, 'customerCode', customer_code,
                                       'pfmeaMethod', pfmea_method, 'cpFormat', cp_format) FROM pkg),
  'document', (SELECT json_build_object('id', d.id, 'docType', d.doc_type, 'methodology', d.methodology,
                                        'docNo', d.doc_no, 'revision', d.revision, 'header', d.header,
                                        'version', d.version)
               FROM documents d WHERE d.package_id = $1 AND d.doc_type = 'PFMEA'),
  'contentVersion', (SELECT content_version FROM pkg),
  'rpnActionThreshold', (SELECT rpn_action_threshold FROM pkg),
  'templatePolicy', (
    SELECT coalesce(jsonb_object_agg(
             CASE t.tbl WHEN 'process_steps' THEN 'processStep' WHEN 'step_flows' THEN 'stepFlow'
                        WHEN 'characteristics' THEN 'characteristic' WHEN 'failure_modes' THEN 'failureMode'
                        WHEN 'failure_effects' THEN 'failureEffect' WHEN 'failure_causes' THEN 'failureCause'
                        WHEN 'failure_chains' THEN 'failureChain' WHEN 'controls' THEN 'control'
                        WHEN 'cp_lines' THEN 'cpLine' WHEN 'reaction_plans' THEN 'reactionPlan' ELSE t.tbl END,
             (SELECT jsonb_agg(snake_to_camel(col)) FROM jsonb_array_elements_text(t.cols) AS col)), '{}'::jsonb)
    FROM pkg JOIN packages g ON g.id = pkg.policy_package_id
    CROSS JOIN LATERAL jsonb_each(g.override_policy) AS t(tbl, cols)),
  'scSymbols', (
    SELECT coalesce(json_agg(json_build_object('id', s.id, 'code', s.code, 'name', s.name, 'isCritical', s.is_critical,
                                               'customerSymbol', m.customer_symbol) ORDER BY s.sort_order, s.code), '[]')
    FROM sc_symbols s
    LEFT JOIN customer_sc_symbols m ON m.sc_symbol_id = s.id AND m.customer_id = (SELECT customer_id FROM pkg)),
  'steps', (
    SELECT coalesce(json_agg(json_build_object(
             'id', t.id, 'opNo', t.op_no, 'seq', t.seq, 'name', t.name, 'function', t.function, 'symbol', t.symbol,
             'kind', t.kind, 'isOptional', t.is_optional, 'machines', t.machines, 'inputs', t.inputs,
             'outputs', t.outputs, 'department', t.department, 'wiRef', t.wi_ref, 'notAnalyzed', t.not_analyzed,
             'notAnalyzedReason', t.not_analyzed_reason, 'generalMode', t.general_mode,
             'origin', t.origin, 'syncStatus', t.sync_status, 'sourceRev', t.source_rev,
             'overrides', (SELECT coalesce(jsonb_object_agg(snake_to_camel(k), v), '{}') FROM jsonb_each(t.overrides) e(k, v)),
             'detachReason', t.detach_reason, 'version', t.version) ORDER BY t.seq), '[]')
    FROM process_steps t WHERE t.package_id = $1),
  'characteristics', (
    SELECT coalesce(json_agg(json_build_object(
             'id', t.id, 'stepId', t.step_id, 'charNo', t.char_no, 'kind', t.kind, 'name', t.name, 'spec', t.spec,
             'lsl', t.lsl, 'usl', t.usl, 'unit', t.unit, 'scSymbolId', t.sc_symbol_id, 'seq', t.seq,
             'origin', t.origin, 'syncStatus', t.sync_status, 'sourceRev', t.source_rev,
             'overrides', (SELECT coalesce(jsonb_object_agg(snake_to_camel(k), v), '{}') FROM jsonb_each(t.overrides) e(k, v)),
             'detachReason', t.detach_reason, 'version', t.version) ORDER BY t.seq, t.char_no), '[]')
    FROM characteristics t WHERE t.package_id = $1),
  'failureModes', (
    SELECT coalesce(json_agg(json_build_object(
             'id', t.id, 'stepId', t.step_id, 'characteristicId', t.characteristic_id, 'text', t.text, 'seq', t.seq,
             'origin', t.origin, 'syncStatus', t.sync_status, 'sourceRev', t.source_rev,
             'overrides', (SELECT coalesce(jsonb_object_agg(snake_to_camel(k), v), '{}') FROM jsonb_each(t.overrides) e(k, v)),
             'detachReason', t.detach_reason, 'version', t.version) ORDER BY t.seq, t.id), '[]')
    FROM failure_modes t WHERE t.package_id = $1),
  'failureEffects', (
    SELECT coalesce(json_agg(json_build_object(
             'id', t.id, 'failureModeId', t.failure_mode_id, 'level', t.level, 'text', t.text, 's', t.s, 'seq', t.seq,
             'origin', t.origin, 'syncStatus', t.sync_status, 'sourceRev', t.source_rev,
             'overrides', (SELECT coalesce(jsonb_object_agg(snake_to_camel(k), v), '{}') FROM jsonb_each(t.overrides) e(k, v)),
             'detachReason', t.detach_reason, 'version', t.version) ORDER BY t.seq, t.id), '[]')
    FROM failure_effects t WHERE t.package_id = $1),
  'failureCauses', (
    SELECT coalesce(json_agg(json_build_object(
             'id', t.id, 'failureModeId', t.failure_mode_id, 'text', t.text, 'seq', t.seq,
             'origin', t.origin, 'syncStatus', t.sync_status, 'sourceRev', t.source_rev,
             'overrides', (SELECT coalesce(jsonb_object_agg(snake_to_camel(k), v), '{}') FROM jsonb_each(t.overrides) e(k, v)),
             'detachReason', t.detach_reason, 'version', t.version) ORDER BY t.seq, t.id), '[]')
    FROM failure_causes t WHERE t.package_id = $1),
  'failureChains', (
    SELECT coalesce(json_agg(json_build_object(
             'id', t.id, 'failureModeId', t.failure_mode_id, 'failureCauseId', t.failure_cause_id,
             's', t.s, 'o', t.o, 'd', t.d, 'rpn', t.rpn, 'justification', t.justification, 'oEvidence', t.o_evidence,
             'seq', t.seq, 'origin', t.origin, 'syncStatus', t.sync_status, 'sourceRev', t.source_rev,
             'overrides', (SELECT coalesce(jsonb_object_agg(snake_to_camel(k), v), '{}') FROM jsonb_each(t.overrides) e(k, v)),
             'detachReason', t.detach_reason, 'version', t.version) ORDER BY t.seq, t.id), '[]')
    FROM failure_chains t WHERE t.package_id = $1),
  'controls', (
    SELECT coalesce(json_agg(json_build_object(
             'id', t.id, 'failureChainId', t.failure_chain_id, 'kind', t.kind, 'text', t.text,
             'controlLibraryId', t.control_library_id, 'isSystemControl', t.is_system_control,
             'systemControlReason', t.system_control_reason, 'seq', t.seq,
             'origin', t.origin, 'syncStatus', t.sync_status, 'sourceRev', t.source_rev,
             'overrides', (SELECT coalesce(jsonb_object_agg(snake_to_camel(k), v), '{}') FROM jsonb_each(t.overrides) e(k, v)),
             'detachReason', t.detach_reason, 'version', t.version) ORDER BY t.seq, t.id), '[]')
    FROM controls t WHERE t.package_id = $1),
  'actions', (
    SELECT coalesce(json_agg(json_build_object(
             'id', t.id, 'failureChainId', t.failure_chain_id, 'kind', t.kind, 'text', t.text,
             'responsibleUserId', t.responsible_user_id, 'responsibleText', t.responsible_text,
             'targetDate', t.target_date, 'status', t.status, 'actionTaken', t.action_taken,
             'completedOn', t.completed_on, 'newS', t.new_s, 'newO', t.new_o, 'newD', t.new_d, 'newRpn', t.new_rpn,
             'designChangeNote', t.design_change_note, 'seq', t.seq, 'version', t.version) ORDER BY t.seq, t.id), '[]')
    FROM actions t WHERE t.package_id = $1)
)
