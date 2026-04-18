-- 016: Restore empresa contact phone column name
ALTER TABLE empresas CHANGE COLUMN telefono_contacto telefono VARCHAR(30) NULL;
