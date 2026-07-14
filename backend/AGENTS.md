Coding assistants for Go backend API development using GoFiber + Huma

## Overview
Three specialized agents work in tandem to help you **write**, **test**, and **profile** production-grade code. Each agent assumes intermediate Go knowledge and enforces best practices specific to the stack.
Make sure to design with clean architecture guidelines.

### Project Layout
```
backend/
├─ apps/
│  ├─ api/                      # HTTP handlers, routes, and presenters
│  │  ├─ handler/               # HTTP handler
│  │  │                         # 
│  │  ├─ middleware/            # Fiber middleware
│  │  │                         # 
│  │  ├─ presenter/             # request and response struct
│  │  │                         # 
│  │  ├─ routes/                # Register handler here
│  │  │                         # Also include the Huma's OpenAPI spec
│  │  │                         # 
│  │  ├─ schema/                # Huma Operation schema
│  │  │                         # Will be included when registering `handlers` to `routes`
│  │  └─ api.go                 # Fiber App object
│  ├─ cmd/                      # Command line interface package
│  │  │                         # Ignore this
│  │  └─ cmd.go                 # The Commandline App object
│  ├─ workers/                  # Background task queue worker
│  │  └─ workers.go             # AsynqWorker that wraps asynq.Server and external 
│  │  │                         # app connection/client such as Gorm DB, SMTP, etc
│  │  ├─ handlers/              # asynq handlers
├─ extensions/                  # Third party extension such as DB, Redis, etc
│  ├─ database/   
│  │  └─ database.go            # gorm.DB
│  ├─ database/   
│  │  └─ email.go               # SMTP client
├─ internal/                    #
│  └─ config.go                 # Configuration
├─ pkg/                         # Core business logic and entities
│  ├─ entities/                 # Pure business entities
│  │  └─ auth.go                # The Commandline App object
│  │  └─ user.go                # The Commandline App object
│  ├─ auth/                     # repository + services for auth entity
│  ├─ user/                     # repository + services for user entity
.air.toml                       # `air` configuration
go.mod
main.go                         # the main func
Makefile                        # Make file
```

### Layes Reponsibilities
| Layer | Contains | Should never import |
|-------|----------|---------------------|
| api/handler | Endpoint handle | |
| api/routes | Fiber route functions, Huma operation registration | `pkg/entities*`, `gorm.io/*` |
| api/presenter | API Request + response | `gorm.op/*` |
| api/schema | OpenAPI (using Huma) operation for handlers | `gorm.op/*` |
| pkg/  | Business logic | `apps/*`, `github.com/gofiber/*`, `github.com/danielgtaylor/huma` |
| pkg/entities | Entities (Core Business Logic) | `apps/*`/ , `pkg/<feature_name>/*`/ |
| pkg/<feature_name> | Repository (Entities CRUD) and Use case (application logic)  | `apps/*`/ |


## Agent Spesific Guidelines
1. Never place business logic in the `apps`` package.
2. All imports must respect the layer boundaries (see the table above).
3. Use constructor injection (NewRepository(db), NewService(repository), NewUserHandler(svc))
4. Return domain errors from services; handlers translate them to HTTP status codes.
5. Keep DTOs (`apps/api/presenter`) JSON‑only – agents should not add gorm tags here.
6. When generating new feature, start with
   - `pkg/entities/<feature_name>.go`: struct of DB model (if needed)
   - `apps/api/presenter/<feature_name>.go`: request/response
   - `pkg/<feature_name>/repository.go`: reposotory method (entities CRUD)
   - `pkg/<feature_name>/service.go`: Use case (application logic) of the entities
   - `apps/api/handler/<feature_name>.go`: handler
   - `apps/api/schema/<feature_name>.go`: Huma operation
   - `apps/api/routes/<feature_name>.go`: Registration of handler to the API

Following the order above ensures the agent never produces “circular” imports.
