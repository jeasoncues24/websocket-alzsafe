# Story S-6.3: Generación JWT por Empresa

## Epic
Epic 6: Sistema de Autenticación JWT por Empresa

## Prioridad
P0

## Estado
pending

## Overview

Implementar función para generar JWT con claims: `empresa_id`, `nombre`, `permissions`, `version`, `phones`.

## Acceptance Criteria

- [ ] Función `GenerateEmpresaJWT(empresa)` retorna string JWT
- [ ] Claims: sub, nombre, perms, ver, exp (5 años), phones
- [ ] Tests unitarios pasando

## Código

```go
func GenerateEmpresaJWT(empresa *Empresa, telefonos []*Telefono) (string, error) {
    now := time.Now()
    exp := now.Add(5 * 365 * 24 * time.Hour) // 5 años

    phoneIDs := make([]string, len(telefonos))
    for i, t := range telefonos {
        phoneIDs[i] = strconv.FormatInt(t.ID, 10)
    }

    claims := JWTClaims{
        EmpresaID:    strconv.FormatInt(empresa.ID, 10),
        Nombre:     empresa.Nombre,
        Permissions: empresa.Permissions,
        Version:    empresa.TokenVersion,
        Telefonos:  phoneIDs,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: exp.Unix(),
            IssuedAt:  now.Unix(),
            Id:        uuid.New().String(),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(secretKey)
}
```

## Dependencies
- S-6.2 (modelos)

## Estimated Effort
1 day