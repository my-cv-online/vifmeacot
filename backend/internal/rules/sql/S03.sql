-- rule: S03
-- reads: cp_lines, characteristics, controls
-- message: CP line {{.charNo}} ({{.symbol}}, step {{.opNo}}) has no strong control: SPC, 100% automated or error-proofing.
-- returns: object_type, object_id, field, key, params (lihat docs/06-rules.md)
SELECT 'cp_lines'::text, l.id, 'controlMethod'::text, ''::text,
       jsonb_build_object('charNo', c.char_no, 'symbol', sc.code, 'opNo', s.op_no)
FROM cp_lines l
JOIN characteristics c ON c.id = l.characteristic_id
JOIN sc_symbols sc ON sc.id = c.sc_symbol_id
JOIN process_steps s ON s.id = l.step_id
LEFT JOIN controls k ON k.id = l.control_id
LEFT JOIN control_library lib ON lib.id = k.control_library_id
WHERE l.package_id = @package_id
  AND NOT l.is_error_proofing
  AND NOT coalesce(lib.is_strong, false)
  AND NOT coalesce(lib.is_error_proofing, false)
