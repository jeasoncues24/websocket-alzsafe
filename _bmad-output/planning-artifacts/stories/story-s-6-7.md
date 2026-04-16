# Story S-6.7: Tests Unitarios

## Epic
Epic 6: Sistema de Autenticación JWT por Empresa

## Prioridad
P0

## Estado
pending

## Overview

Implementar tests unitarios para el sistema JWT y modelos.

## Coverage

- [ ] GenerateEmpresaJWT: ≥ 90%
- [ ] ValidateJWT: ≥ 90%
- [ ] Empresa model: ≥ 80%
- [ ] Telefono model: ≥ 80%

## Casos de Prueba

```go
func TestGenerateEmpresaJWT(t *testing.T) {
    // Happy path
    empresa := &Empresa{ID: 1, RUC: "20123456789", Nombre: "Test", TokenVersion: 1, Permissions: []string{"send"}}
    telefonos := []*Telefono{{ID: 1, PhoneNumber: "+519999999999", Status: "active"}}
    
    token, err := GenerateEmpresaJWT(empresa, telefonos)
    assert.NoError(t, err)
    assert.NotEmpty(t, token)
}

func TestValidateJWT(t *testing.T) {
    tests := []struct {
        name    string
        token  string
        errMsg string
    }{
        {"valid", validToken, ""},
        {"expired", expiredToken, "token expired"},
        {"invalid_signature", badSigToken, "invalid signature"},
        {"revoked", revokedToken, "token revoked"},
    }
}
```

## Dependencias
- S-6.3, S-6.4

## Estimated Effort
2 days