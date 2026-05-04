-- 016: Revertir seeds iniciales en orden inverso de dependencias
DELETE FROM user_modules WHERE user_id = 1;
DELETE FROM admin_users WHERE username = 'admin_usqay';
DELETE FROM modules WHERE slug IN ('dashboard', 'companies', 'users', 'roles', 'modules', 'sessions', 'messages', 'broadcasts');
DELETE FROM empresas WHERE ruc = '20100000001';
DELETE FROM roles WHERE name IN ('super_admin', 'admin', 'operador', 'viewer');
