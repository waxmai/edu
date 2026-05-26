-- rollback-policy: forward-only
-- rollback-note: Seed data is idempotent; adjust with a new bootstrap/update migration.
-- Demo-only reference rows. Do not replace these with real personal data in source control.
INSERT INTO `admin` (`id`, `username`, `mobile`) VALUES
(1, 'Demo Admin A', '10000000001'),
(2, 'Demo Admin B', '10000000002'),
(3, 'Demo Admin C', '10000000003')
ON DUPLICATE KEY UPDATE username = VALUES(username), mobile = VALUES(mobile);
