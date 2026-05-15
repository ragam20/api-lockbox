# Project Overview

This project is a secure vault-style authentication system using client-side cryptography.
Users never send raw passwords to the backend. The frontend derives cryptographic materials and only sends verifiers, salts, and encrypted data.

---

# Tech Stack

- Golang
- Redis
- Gin (web framework)
- PostgreSQL
- SMTP (gomail)
- Viper for environment management
- JWT for session/authentication

---

# Project Structure

```txt id="jlwm4x"
.
├── cmd
│   └── api
└── internal
    ├── app
    │   └── auth
    ├── cache
    ├── config
    ├── constant
    ├── database
    │   ├── entity
    │   └── repository
    ├── http
    │   ├── handler
    │   ├── middleware
    │   ├── request
    │   ├── response
    │   ├── router
    │   └── server
    ├── idgen
    ├── logger
    └── smtp
```

---

#### cmd/api

This directory is used for:

- application entry point
- dependency injection

---

#### internal

This directory is used for managing internal application dependencies such as:

- database
- Redis
- application services
- internal packages

This package should only be used internally by the application.

---

#### internal/app

This directory is used for managing business logic.

If you want to add a new feature, create a new directory inside this package.

Example:

```txt id="jlwmbk"
internal/app/auth
```

---

#### internal/app/(feature)

Every feature should contain:

- repository interface
- service.go (business logic)
- domain model

---

#### internal/config

This directory is used for managing environment configuration.

---

#### internal/constant

This directory is used for managing constant values.

---

#### internal/database

This directory is used for managing:

- database configuration
- entities
- repositories

---

#### internal/http

This directory is used for managing HTTP/server-related logic.

This project uses Gin as the web framework.

Includes:

- handlers
- middleware
- request DTOs
- response DTOs
- routers
- server setup

---

# Important Constraints

- Do not install unnecessary dependencies
- Do not modify unrelated files
- Keep code simple and maintainable
- Prefer explicit code over magic abstractions
- Use context timeout for external resources

---

# JSON Conventions

Use camelCase for all JSON request and response bodies.

Example:

```json id="jlwm8c"
{
  "authVerifier": "",
  "masterKeySalt": ""
}
```

---

# API Rules

Use structured JSON responses.

## Success Response

```json id="0jlwmm"
{
  "success": true,
  "message": "Success",
  "timestamp": "",
  "data": {}
}
```

## Error Response

```json id="jlwmrw"
{
  "success": false,
  "message": "Invalid token",
  "timestamp": ""
}
```

Use proper HTTP status codes.
