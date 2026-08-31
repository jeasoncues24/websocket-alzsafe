-- 016: Seeds iniciales consolidados
-- INSERT IGNORE en todos los bloques: idempotente si los registros ya existen (UNIQUE key silencia el duplicado).

INSERT IGNORE INTO roles (name, description, is_root, permissions, created_by) VALUES
('admin',               'Administrador',        TRUE,  '["all"]',                                                         NULL),
('soporte usqay',       'Soporte Usqay',        FALSE, '["companies","messages","sessions","broadcasts"]',                 NULL),
('administracion usqay','Administración Usqay', FALSE, '["all"]',                                                         NULL);

-- ── Módulos ──────────────────────────────────────────────────────────────────
INSERT IGNORE INTO modules (name, slug, description) VALUES
('dashboard',  'dashboard',  'Panel de control'),
('companies',  'companies',  'Gestión de empresas'),
('users',      'users',      'Gestión de usuarios'),
('roles',      'roles',      'Gestión de roles'),
('modules',    'modules',    'Gestión de módulos'),
('sessions',   'sessions',   'Sesiones WhatsApp'),
('messages',   'messages',   'Mensajes'),
('broadcasts', 'broadcasts', 'Difusiones');

-- ── Admin (fulanito) ──────────────────────────────────────────────────────────
INSERT IGNORE INTO admin_users (username, password_hash, email, role_id, activo, created_by)
SELECT 'fulanito', '$2a$12$qP0VyYTQiMkrijBCQeJUuujabI587TsTQQjdSv8AXypReqFLS/gkK', 'fulanito@usqay.pe', id, TRUE, NULL
FROM roles WHERE name = 'admin';

-- ── Soporte Usqay (5 usuarios) ───────────────────────────────────────────────
-- contraseña = username de cada uno
INSERT IGNORE INTO admin_users (username, password_hash, email, role_id, activo, created_by)
SELECT u.username, u.password_hash, u.email, r.id, TRUE, NULL
FROM (
    SELECT 'lsabogal'        AS username, '$2a$12$J9/dzuE9TqzysSyr6/4Q7OVcxMgtUrOGrtsySHCHnCye.IQ65d7sm' AS password_hash, 'lgutierrez@usqay.pe'    AS email UNION ALL
    SELECT 'pedrovall',                   '$2a$12$e.e5abiz7086740E7ICFneH5NYq8I.2T0vF3WOTUcDmaj2vGDN8Y2',                 'pvalladares@usqay.pe'   UNION ALL
    SELECT 'luisillosupport',             '$2a$12$gbgTcU2F9kopMLAejPMmIuAZANKO26cHJbyDoyA7sXI8YdfUr3I0W',                 'langel@usqay.pe'        UNION ALL
    SELECT 'lejzer',                      '$2a$12$DTvdokLzGYMO5JtNJhBMVe2ajgS0BblVENFV2Enik7dLERLgkBzBK',                 'lparedes@usqay.pe'      UNION ALL
    SELECT 'arnoldsupport',               '$2a$12$n/QLDkrH4bBSHPq6ITbCoOkO4e.PONMODCzcT6hAFZ1YcZLf2v3hm',                 'acastillo@usqay.pe'
) u CROSS JOIN roles r WHERE r.name = 'soporte usqay';

-- ── Administración Usqay (2 usuarios) ────────────────────────────────────────
INSERT IGNORE INTO admin_users (username, password_hash, email, role_id, activo, created_by)
SELECT u.username, u.password_hash, u.email, r.id, TRUE, NULL
FROM (
    SELECT 'toñoadmin'       AS username, '$2a$12$xmuRxpxY.21tg1PwzIf8QevtLZJwui65nNwnJg6BaPkb2uNr1o.hC' AS password_hash, 'gdelaselva@usqay.pe'    AS email UNION ALL
    SELECT 'enriquezapata',               '$2a$12$h1InbnSnr3K1itaalXRt1upcwOLv3Yl96plSjr6FK1Rvg03iFqg0G',                 'ezapata@usqay.pe'
) u CROSS JOIN roles r WHERE r.name = 'administracion usqay';

-- ── Módulos por rol ───────────────────────────────────────────────────────────
-- admin: todos los módulos
INSERT IGNORE INTO user_modules (user_id, module_id)
SELECT u.id, m.id FROM admin_users u CROSS JOIN modules m
WHERE u.username = 'fulanito';

-- soporte: solo companies, messages, sessions, broadcasts
INSERT IGNORE INTO user_modules (user_id, module_id)
SELECT u.id, m.id FROM admin_users u CROSS JOIN modules m
WHERE u.username IN ('lsabogal', 'pedrovall', 'luisillosupport', 'lejzer', 'arnoldsupport')
  AND m.slug IN ('companies', 'messages', 'sessions', 'broadcasts');

-- administración: todos los módulos
INSERT IGNORE INTO user_modules (user_id, module_id)
SELECT u.id, m.id FROM admin_users u CROSS JOIN modules m
WHERE u.username IN ('toñoadmin', 'enriquezapata');
