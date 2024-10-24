# Password Keeper
Keeper for private data on Golang

### Installation
```go
go build main.go
./main.exe --config=/path/to/config.json
```

In config file you should specify DSN to postgres database file and port for gRPC server.  
```
{
    "database_dsn": "postgres://postgres:postgres@localhost:5432/postgres",
    "port": 4040
}
```

Сервер реализовывает следующую бизнес-логику:
  - регистрация, аутентификация и авторизация пользователей;
  - хранение приватных данных;
  - синхронизация данных между несколькими авторизованными клиентами одного владельца;
  - передача приватных данных владельцу по запросу.
  
Клиент реализовывает следующую бизнес-логику:
  - аутентификация и авторизация пользователей на удалённом сервере;
  - доступ к приватным данным по запросу.

Приложение реализовано как TUI. Взаимодействие происходит по gRPC сервису.
