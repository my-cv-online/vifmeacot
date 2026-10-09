-- rule: K06
-- reads: documents, packages, parts
-- message: {{.docType}} header uses part {{.docPartNo}} / change level {{.docChangeLevel}}, but the package part is {{.partNo}} / {{.changeLevel}}.
-- returns: object_type, object_id, field, key, params (lihat docs/06-rules.md)
SELECT 'documents'::text, d.id, 'header'::text, d.doc_type::text,
       jsonb_build_object('docType', d.doc_type, 'docPartNo', coalesce(d.header ->> 'partNo', '-'),
                          'docChangeLevel', coalesce(d.header ->> 'changeLevel', '-'),
                          'partNo', p.part_no, 'changeLevel', coalesce(p.change_level, '-'))
FROM documents d
JOIN packages pk ON pk.id = d.package_id
JOIN parts p ON p.id = pk.part_id
WHERE d.package_id = @package_id
  AND ((d.header ? 'partNo' AND d.header ->> 'partNo' IS DISTINCT FROM p.part_no)
    OR (d.header ? 'changeLevel' AND d.header ->> 'changeLevel' IS DISTINCT FROM p.change_level))
