# Go Start Here

> ⚠️ **This project is under heavy development.** Breaking changes may occur without notice. Use at your own risk in production environments.

## About

This is fullstack web application, built with Go (Fiber + Huma) for backend and Typescript (Svelte) for frontend.


## URL

### Dev Mode

- [Backend: API Doc](http://127.0.0.1:8080/docs)
- [Frontend](http://127.0.0.1:8080)
- [MailHog](http://127.0.0.1:8025)
- [Traefik Dashboard](http://127.0.0.1:8081/dashboard/)

## Features

### Backend

- Rest API
- OpenAPI documentation
- Background task queue worker


### Frontend

- OAuth2 for easy authentication


## Configuration

### `.env`

### `backend/.env`

### `frontenv/.env`


## Tools

### Backend


### Frontend


## Usage

### Development

```sh
docker compose --env-file .env --env-file backend/.env --env-file frontend/.env \
    -f docker-compose.base.yml -f docker-compose.dev.yml  \
    up
```


Includes Mailhog for local email testing (SMTP at `localhost:1025`, UI at `http://localhost:8025`).


### Production

```sh
docker compose --env-file .env --env-file backend/.env --env-file frontend/.env \
    -f docker-compose.base.yml -f docker-compose.prod.yml  \
    up
```


## References and Inspirations

- [github.com/fastapi/full-stack-fastapi-template](https://github.com/fastapi/full-stack-fastapi-template): Full stack, modern web application template. Using FastAPI, React, SQLModel, PostgreSQL, Docker, GitHub Actions, automatic HTTPS and more.


[asynq]: https://github.com/hibiken/asynq
[fiber]: https://docs.gofiber.io
[huma]: https://huma.rocks
[svelte]: https://svelte.dev/docs/svelte
[sveltekit]: https://svelte.dev/docs/kit
[shacdn-svelte]: https://shadcn-svelte.com
