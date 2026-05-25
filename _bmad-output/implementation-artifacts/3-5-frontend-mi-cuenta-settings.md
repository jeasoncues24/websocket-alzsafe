# Story 3-5: Frontend — Sección "Mi cuenta" en Settings

**Estado:** review
**Epic padre:** Epic 3 — Módulos Dinámicos desde BD y Perfil de Usuario
**Story ID:** 3.5

---

## Story

Como **Usuario Autenticado**,
quiero **contar con una pestaña dedicada llamada "Mi cuenta" dentro de la pantalla de Configuración**,
para **poder visualizar y actualizar mis datos personales de perfil y mi contraseña con una interfaz premium e intuitiva**.

---

## Acceptance Criteria

**AC1 — Pestaña "Mi cuenta" integrada en `/settings`**
Dado un usuario que navega a la sección de Configuración (`/settings`),
cuando se renderiza la página,
entonces se visualiza la pestaña "Mi cuenta" (o "Perfil") junto a las existentes ("General", "Apariencia", "Acerca de").

**AC2 — Formulario de Datos Personales funcional con feedback visual**
Dado el formulario "Datos Personales" en la pestaña "Mi Cuenta",
cuando el usuario ingresa a la sección:
- Los campos "Nombre de Usuario" e "Email" se pre-cargan automáticamente con la información del usuario autenticado actualmente (obtenida del store).
- Al hacer modificaciones válidas y presionar "Guardar Cambios", se realiza la llamada HTTP `PUT /api/auth/me`.
- Durante el envío del formulario, el botón de guardado muestra un spinner de carga (`lucide-react/Loader2` u otro) y se deshabilita temporalmente.
- Tras la actualización exitosa, se actualiza el store de Zustand local, se muestra un Toast verde de éxito y se desactivan los estados de edición. Si hay error (ej. email duplicado), se muestra un Toast rojo descriptivo.

**AC3 — Formulario de Cambio de Contraseña seguro**
Dado el formulario "Cambiar Contraseña" en la pestaña "Mi Cuenta",
cuando el usuario ingresa a la sección:
- Visualiza los campos "Contraseña Actual", "Nueva Contraseña" y "Confirmar Contraseña" (todos de tipo `password` con opción visual de alternar visibilidad si el diseño general lo maneja).
- El sistema realiza validaciones del lado del cliente (ej. la nueva contraseña y su confirmación deben ser idénticas, no estar vacías y cumplir con longitud mínima).
- Al presionar "Actualizar Contraseña", se realiza la llamada HTTP `PUT /api/auth/me/password`.
- Durante el envío, el botón muestra feedback de carga.
- Al completarse con éxito, se muestra un Toast de confirmación y se limpian por completo todos los campos del formulario.

**AC4 — Diseño Premium (UX Pro Max)**
La interfaz visual de la pestaña cumple con los más altos estándares modernos:
- Estructura limpia y balanceada (layouts adaptativos que lucen perfectos en desktop y móvil).
- Uso de componentes locales y estilos consistentes (inputs con focos sutiles, bordes redondeados y colores HSL coherentes con el tema oscuro/claro).
- Micro-animaciones en botones y estados activos/inactivos para que la interfaz se sienta fluida e interactiva.

**AC5 — Compilación y Buenas Prácticas**
`cd frontend && npm run build` y `cd frontend && npm run lint` completan su ejecución exitosamente sin warnings ni errores de TypeScript.

---

## Tasks / Subtasks

- [ ] **T1 — Integrar la pestaña en la UI de configuraciones**
  - **Archivo:** `frontend/app/settings/page.tsx`
  - [ ] Añadir la nueva pestaña `"account"` (Mi Cuenta) a los componentes de Tabs (`TabsList`, `TabsTrigger`, `TabsContent`) existentes en la página.
  - [ ] Diseñar el contenedor principal de la pestaña con un espaciado armónico, títulos explicativos claros y divisores visuales limpios.

- [ ] **T2 — Diseñar e implementar el Formulario de Datos de Perfil**
  - **Archivo:** Componente en `frontend/app/settings/page.tsx` (o un subcomponente dedicado)
  - [ ] Añadir inputs para `username` e `email`.
  - [ ] Conectar los inputs con un estado React (`useState`) pre-cargado desde el store `useAppStore`.
  - [ ] Validar los campos en cliente (correo electrónico válido y campos no vacíos) mostrando textos de error en color rojo sutil bajo el input si no son válidos.
  - [ ] Implementar la llamada HTTP de guardado hacia `PUT /api/auth/me`.
  - [ ] Actualizar Zustand al recibir respuesta exitosa y gatillar el Toast de éxito.

- [ ] **T3 — Diseñar e implementar el Formulario de Cambio de Contraseña**
  - **Archivo:** Componente en `frontend/app/settings/page.tsx` (o subcomponente dedicado)
  - [ ] Añadir inputs para `currentPassword`, `newPassword` y `confirmPassword`.
  - [ ] Implementar validaciones visuales locales en tiempo real (ej. si `newPassword !== confirmPassword`, deshabilitar botón y mostrar advertencia).
  - [ ] Implementar la llamada HTTP de actualización hacia `PUT /api/auth/me/password`.
  - [ ] Limpiar los campos del formulario tras un Toast exitoso.

- [ ] **T4 — Pulido responsivo y micro-interacciones**
  - **Archivo:** Hojas de estilos locales o Tailwind classes en los formularios
  - [ ] Ajustar la grilla de formularios para pantallas de escritorio (ej. dos columnas) y colapsar a una sola columna en pantallas móviles.
  - [ ] Añadir micro-transiciones (hover, active, focus) en todos los inputs y botones para dar feedback instantáneo al usuario.

---

## Dev Notes

- **Toasts:** Utilizar el sistema de notificaciones/toasts existente en el proyecto frontend para desplegar los mensajes de éxito y error.
- **Tipado Next.js:** Asegurarse de que no haya variables sueltas ni tipados `any` innecesarios que causen fallas al correr el build de producción.

---

## Dev Agent Record

### Agent Model Used
Gemini 1.5 Pro (Antigravity Coordinator)

### Debug Log References

### Completion Notes List

### File List
