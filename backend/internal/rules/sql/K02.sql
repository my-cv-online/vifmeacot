-- rule: K02
-- reads: characteristics, cp_lines
-- message: Characteristic {{.charNo}} {{.charName}} (step {{.opNo}}) has no Control Plan line.
-- returns: object_type, object_id, field, key, params (lihat docs/06-rules.md)
SELECT 'characteristics'::text, c.id, 'name'::text, ''::text,
       jsonb_build_object('charNo', c.char_no, 'charName', c.name, 'opNo', s.op_no)
FROM characteristics c
JOIN process_steps s ON s.id = c.step_id
WHERE c.package_id = @package_id
  AND c.sc_symbol_id IS NULL                   -- special characteristic dicek oleh S01
  AND NOT EXISTS (SELECT 1 FROM cp_lines l WHERE l.package_id = c.package_id AND l.characteristic_id = c.id)
