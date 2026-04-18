# Story 7.1: Login con token JWT

Status: in-progress

## Story

As a empresa user,
I want to login with my JWT token,
so that I can access my company panel without username/password.

## Acceptance Criteria

1. [AC: login-jwt] Login page accepts JWT token via URL query param
2. [AC: login-redirect] Valid token redirects to dashboard
3. [AC: login-invalid] Invalid/expired token shows error

## Flow

1. User visits `/empresa/login?token=xxx`
2. Backend validates JWT
3. On success → redirect to `/empresa/dashboard`
4. On failure → show error message

## Implementation

- New route: `/empresa/login`
- Validate token via empresaAuthMiddleware.ParseToken
- Set session/cookie on success
- Redirect to dashboard