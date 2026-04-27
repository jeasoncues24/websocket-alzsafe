-- 016: Rename empresa contact phone column
ALTER TABLE empresas CHANGE COLUMN telefono telefono_contacto VARCHAR(30) NULL;
