export interface User {
    id?: number;
    razon_social: string;
    ruc: string;
    nombre_comercial: string;
    telefono: string;
    codigo_postal: string;
    is_active?: number;
    is_linked?: number;
    created_at?: Date;
}