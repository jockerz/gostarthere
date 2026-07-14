# Output

- Openapi.json: http://127.0.0.1:8080/openapi.json
- API Documentation: http://127.0.0.1:8080/docs


# Run Service

## API

```sh
make run api
# OR
go run main.go api
```

## SMTP Server

For development, use Python `smptd`: `python3 -m smtpd -n -c DebuggingServer localhost:1578`


## Worker

```sh
make run worker
# OR
go run main.go worker
```


## Test

```sh
make test
# OR 
go test ./...
```
