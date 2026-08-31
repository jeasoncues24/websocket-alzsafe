-- 016: Revertir seeds iniciales en orden inverso de dependencias FK.
DELETE FROM user_modules WHERE user_id IN (
    SELECT id FROM admin_users WHERE username IN (
        'fulanito', 'lsabogal', 'pedrovall', 'luisillosupport',
        'lejzer', 'arnoldsupport', 'toñoadmin', 'enriquezapata'
    )
);
DELETE FROM admin_users WHERE username IN (
    'fulanito', 'lsabogal', 'pedrovall', 'luisillosupport',
    'lejzer', 'arnoldsupport', 'toñoadmin', 'enriquezapata'
);
DELETE FROM modules     WHERE slug IN ('dashboard','companies','users','roles','modules','sessions','messages','broadcasts');
DELETE FROM roles       WHERE name IN ('admin','soporte usqay','administracion usqay');
