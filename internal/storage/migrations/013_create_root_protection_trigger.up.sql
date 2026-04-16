-- Migration 013: Create trigger to protect is_root in roles
DELIMITER $$

CREATE TRIGGER prevent_root_role_update
BEFORE UPDATE ON roles
FOR EACH ROW
BEGIN
    IF NEW.is_root != OLD.is_root THEN
        SIGNAL SQLSTATE '45000' 
            SET MESSAGE_TEXT = 'No se puede modificar el flag is_root de un rol root';
    END IF;
END$$

DELIMITER ;