# Go Start Here

> ⚠️ **This project is under heavy development.** Breaking changes may occur without notice. Use at your own risk in production environments.

## About

This is fullstack web application, built with Go (Fiber + Huma) for backend and Typescript (Svelte + Shadcn) for frontend.


## Technology Stack and Features

- [Huma][huma] + [Fiber][fiber] for API web application and API documentation
- [Asynq][asynq] for background task/queue
- [Svelte][svelte] + [Shadcn Svelte][shacdn-svelte] for the frontend
- [Docker Compose](https://docs.docker.com/compose) for local services and self-hosted deployment.
  - [Traefik](https://traefik.io) as a reverse proxy with automatic HTTPS.
- [Mailhog][mailhog]: *development mode* email server

_More_

- Email based account activation and password recovery


## Configuration

### `.env`

### `backend/.env`

### `frontenv/.env`


## Tools

### Backend


### Frontend


## Usage

### Development

For development purposes, this docker compose includes [Mailhog][mailhog]

Start

```sh
docker compose --env-file .env \
    -f docker-compose.base.yml -f docker-compose.dev.yml  \
    up
```

Stop

```sh
docker compose --env-file .env \
    -f docker-compose.base.yml -f docker-compose.dev.yml  \
    down
```

Need to recompile the Go sources or rebuild the svelte code?
Delete the current image for the container.

```sh
docker -f docker-compose.base.yml -f docker-compose.dev.yml  \
    down --rmi local

# Remove Frontend only
docker -f docker-compose.base.yml -f docker-compose.dev.yml  \
    down --rmi local frontend
```


- Frontend: [http://localhost:8080](http://localhost:8080)
- Traefik: [http://localhost:8081](http://localhost:8081)
- Mailhog: [http://localhost:8025](http://localhost:8025)


### Production

⚠️ **The production mode is under heavy development.**

Take note that in production mode, you need to use external SMTP server.


```sh
docker compose --env-file .env \
    -f docker-compose.base.yml -f docker-compose.prod.yml  \
    up
```


## References and Inspirations

- [github.com/fastapi/full-stack-fastapi-template](https://github.com/fastapi/full-stack-fastapi-template): Full stack, modern web application template. Using FastAPI, React, SQLModel, PostgreSQL, Docker, GitHub Actions, automatic HTTPS and more.


[asynq]: https://github.com/hibiken/asynq
[fiber]: https://docs.gofiber.io
[huma]: https://huma.rocks
[mailhog]: https://github.com/mailhog/MailHog
[svelte]: https://svelte.dev/docs/svelte
[sveltekit]: https://svelte.dev/docs/kit
[shacdn-svelte]: https://shadcn-svelte.com
