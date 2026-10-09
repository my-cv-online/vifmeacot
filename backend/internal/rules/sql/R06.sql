-- rule: R06
-- reads: cp_lines
-- message: Error-proofing on CP line {{.charNo}} (step {{.opNo}}) has no verification frequency.
-- returns: object_type, object_id, field, key, params (lihat docs/06-rules.md)
SELECT 'cp_lines'::text, l.id, 'epVerifyFreq'::text, ''::text,
       jsonb_build_object('charNo', c.char_no, 'opNo', s.op_no)
FROM cp_lines l
JOIN process_steps s ON s.id = l.step_id
JOIN characteristics c ON c.id = l.characteristic_id
WHERE l.package_id = @package_id
  AND l.is_error_proofing
  AND coalesce(btrim(l.ep_verify_freq), '') = ''
