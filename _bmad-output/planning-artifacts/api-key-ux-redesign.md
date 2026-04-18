# UX/UI Redesign: API Keys por Teléfono

## Objetivo

Rediseñar la experiencia de creación, visualización, rotación y revocación de API keys para que siga el modelo real del backend: la unidad de gestión es el teléfono WhatsApp, no la empresa.

## Principios

- El secreto se muestra una sola vez.
- El teléfono es el contexto principal de la key.
- La empresa es solo el tenant dueño.
- La pantalla debe sentirse como un panel de producto serio: clara, densa y segura.
- Todo debe construirse con shadcn y el tema nuevo del frontend.

## Estructura de la vista

### 1. Header de contexto

- Nombre del teléfono
- RUC/empresa asociada
- Estado operativo del teléfono
- Badges: activo, QR pendiente, desconectado

### 2. Resumen de API Key

- Estado actual de la key
- Prefijo visible
- Versión/rotación
- Último uso
- Expiración

### 3. Acciones primarias

- Crear API key
- Rotar API key
- Revocar API key
- Copiar secreto

### 4. Uso y auditoría

- Uso diario
- Últimos eventos
- Acciones de auditoría

## Componentes shadcn

- `Card`
- `Badge`
- `Button`
- `Dialog`
- `Tabs`
- `Input`
- `Textarea`
- `Alert`
- `Skeleton`
- `Separator` si hace falta

## Estados críticos

- Sin key creada
- Key creada y visible
- Key rotada
- Key revocada
- Teléfono inactivo
- Error de autorización

## Copy sugerido

- "La API key se muestra una sola vez. Guárdala ahora."
- "Rotar esta key invalida la anterior inmediatamente."
- "Revocar esta key corta el acceso de la integración asociada."

## Criterio de éxito

- La UI no sugiere que la key pertenece a la empresa.
- El flujo de creación/rotación es obvio y seguro.
- El tema nuevo se siente consistente en toda la vista.
