-- fixture: F08 di atas db/seed/demo.sql, paket PS-07, @today = 2026-10-08
-- scenario: Customer B menetapkan ambang RPN 100; chain dengan RPN >= 100 tanpa aksi dilaporkan.
-- expect: [{"object_type": "failure_chains", "field": "rpn", "params": {"rpn": 126}}, {"object_type": "failure_chains", "field": "rpn", "params": {"rpn": 144}}, {"object_type": "failure_chains", "field": "rpn", "params": {"rpn": 105}}, {"object_type": "failure_chains", "field": "rpn", "params": {"rpn": 105}}]
UPDATE customers SET rpn_action_threshold = 100 WHERE code = 'CB';
