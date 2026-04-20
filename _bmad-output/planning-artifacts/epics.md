---
stepsCompleted:
  - validate-prerequisites
inputDocuments:
  - _bmad-output/project-context.md
  - _bmad-output/implementation-artifacts/sprint-status.yaml
  - _bmad-output/implementation-artifacts/future-index.md
  - Makefile
  - docker/go/Dockerfile
  - docker/nginx/default.conf
  - .env.example
  - frontend/.env.example
  - frontend/next.config.ts
  - internal/config/config.go
  - frontend/README.md
---

# wsapi - Epic Breakdown

## Overview

This document provides the complete epic and story breakdown for wsapi, decomposing the requirements from the current project context and deployment constraints into implementable work.

## Requirements Inventory

### Functional Requirements

FR1: Backend must be buildable and runnable as a production Docker image.
FR2: Frontend must be buildable and runnable as a production Docker image.
FR3: Production deployment must be orchestrated with Docker Compose.
FR4: The deployment must expose configurable ports and avoid hardcoded port values in source.
FR5: A script must detect occupied ports and help choose replacement values before deployment.
FR6: The Makefile must provide production build targets for backend and frontend.
FR7: The repository must include documentation for production installation and startup.
FR8: Docker build contexts must exclude generated, secret, and runtime artifacts via `.dockerignore`.
FR9: Git tracking rules must exclude production outputs, local runtime data, and environment files that should not be committed.

### NonFunctional Requirements

NFR1: Deployment configuration must be repeatable and deterministic.
NFR2: Source code must not hardcode localhost, IPs, or ports.
NFR3: Production build outputs must not rely on development-only commands.
NFR4: The port selection workflow must prevent or clearly surface collisions.
NFR5: Installation documentation must be operationally complete and concise.
NFR6: Ignore rules must not hide source or deployment files required to reproduce the build.

### Additional Requirements

- Backend already reads `APP_PORT` from `internal/config/config.go`; production deployment must preserve that contract.
- Frontend rewrites depend on `NEXT_PUBLIC_API_URL`; production compose wiring must provide the backend URL explicitly.
- The existing backend Dockerfile under `docker/go/Dockerfile` can be reused or refined for production.
- The existing nginx config proxies to `127.0.0.1:8080`; if a different topology is selected, that proxy contract must be updated.
- A top-level `docker-compose.yml` is not present yet and must be created.
- Port validation should be done by a shell script so the user can verify occupied ports locally before deployment.
- Root `.dockerignore` must be added so Docker builds do not send `.git`, env files, runtime logs, or frontend build output.
- Root `.gitignore` must be expanded to exclude Docker override files, runtime data, and deployment-only artifacts.

### UX Design Requirements

No UX design document was included for this deployment epic.

### FR Coverage Map

FR1: Epic 1 - Build and run the backend in production Docker
FR2: Epic 1 - Build and run the frontend in production Docker
FR3: Epic 1 - Orchestrate production services with Docker Compose
FR4: Epic 1 - Configure non-hardcoded ports and deployment wiring
FR5: Epic 1 - Detect occupied ports before deployment
FR6: Epic 1 - Add production build targets to Makefile
FR7: Epic 1 - Document production installation and startup
FR8: Epic 1 - Exclude generated and secret files from Docker build contexts
FR9: Epic 1 - Exclude runtime and deployment-only artifacts from Git

## Epic List

### Epic 1: Producción Dockerizada de WSAPI
Permitir despliegue productivo de WSAPI con Docker Compose, backend y frontend empaquetados, puertos configurables, higiene de build/contexto y una guía clara de instalación y operación.
**FRs covered:** FR1, FR2, FR3, FR4, FR5, FR6, FR7, FR8, FR9

## Story Draft Status

La estructura del epic ya está definida. El siguiente paso es descomponerla en historias de implementación.
