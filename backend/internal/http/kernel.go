package http

import "net/http"

type Kernel struct {
	AdminAuth    func(http.Handler) http.Handler
	EmpresaAuth  func(http.Handler) http.Handler
	ClientAuth   func(http.Handler) http.Handler
	ServiceStack func(http.Handler) http.Handler
	Global       []func(http.Handler) http.Handler
}

func identityMiddleware(next http.Handler) http.Handler {
	return next
}

// NewKernel inicializa el stack de middlewares globales y de auth.
func NewKernel(auth AdminAuthProvider, empresaAuth EmpresaAuthProvider, apiKeyAuth ClientAuthProvider, telemetryMW func(http.Handler) http.Handler) *Kernel {
	k := &Kernel{
		AdminAuth:   identityMiddleware,
		EmpresaAuth: identityMiddleware,
		ClientAuth:  identityMiddleware,
		Global: []func(http.Handler) http.Handler{
			CORSMiddleware,
			CorrelationIDMiddleware,
			LoggingMiddleware,
		},
	}
	if auth != nil {
		k.AdminAuth = auth.RequireAuth()
	}
	if empresaAuth != nil {
		k.EmpresaAuth = empresaAuth.RequireEmpresaAuth()
	}
	if apiKeyAuth != nil {
		k.ClientAuth = apiKeyAuth.RequireApiKeyAuth()
	}
	if telemetryMW != nil {
		k.ServiceStack = func(next http.Handler) http.Handler {
			return k.ClientAuth(telemetryMW(next))
		}
	} else {
		k.ServiceStack = k.ClientAuth
	}
	return k
}

// Apply envuelve el handler final con los middlewares globales
func (k *Kernel) Apply(h http.Handler) http.Handler {
	for _, middleware := range k.Global {
		h = middleware(h)
	}
	return h
}

type AdminAuthProvider interface {
	RequireAuth() func(http.Handler) http.Handler
}

type EmpresaAuthProvider interface {
	RequireEmpresaAuth() func(http.Handler) http.Handler
}

type ClientAuthProvider interface {
	RequireApiKeyAuth() func(http.Handler) http.Handler
}
