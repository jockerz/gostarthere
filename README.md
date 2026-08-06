# Go Start Here

> ⚠️ **This project is under heavy development.** Breaking changes may occur without notice. Use at your own risk in production environments.

## About

This is fullstack web application, built with Go (Fiber + Huma) for backend and Typescript (Svelte) for frontend.


## Features

### Backend

- OpenAPI documentation
- Background task queue worker


### Frontend

- OAuth2 for easy authentication


## Tools

### Backend


### Frontend


## Usage

### Development

```sh
docker compose -f docker-compose.base.yml -f docker-compose.dev.yml up
```

Or daemon mode
```sh
docker compose -f docker-compose.base.yml -f docker-compose.dev.yml up -d
```

Includes Mailhog for local email testing (SMTP at `localhost:1025`, UI at `http://localhost:8025`).


### Production

```sh
ACME_EMAIL=you@example.com SMTP_HOST=smtp.example.com docker compose -f docker-compose.base.yml -f docker-compose.prod.yml up
```

Or daemon mode
```sh
ACME_EMAIL=you@example.com SMTP_HOST=smtp.example.com docker compose -f docker-compose.base.yml -f docker-compose.prod.yml up -d
```

Set `SMTP_HOST` (and optionally `SMTP_PORT`, `SMTP_STARTTLS`, `SMTP_USERNAME`, `SMTP_PASSWORD`) in `.env` to point the backend at a real SMTP provider.


## References and Inspirations

- [github.com/fastapi/full-stack-fastapi-template](https://github.com/fastapi/full-stack-fastapi-template): Full stack, modern web application template. Using FastAPI, React, SQLModel, PostgreSQL, Docker, GitHub Actions, automatic HTTPS and more.


[asynq]: https://github.com/hibiken/asynq
[fiber]: https://docs.gofiber.io
[huma]: https://huma.rocks
[svelte]: https://svelte.dev/docs/svelte
[sveltekit]: https://svelte.dev/docs/kit
[shacdn-svelte]: https://shadcn-svelte.com
