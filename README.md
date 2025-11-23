# PR Reviewer Service

Микросервис для автоматического назначения ревьюеров на Pull Requestы внутри команд и управления пользователями / командами.
HTTP API описано в `api/openapi/openapi.yml`, Swagger доступен из браузера.

---

## Продакшен

Проект развернут на сервере и доступен по адресу (До 1 февраля):

[https://dbudin.ru](http://dbudin.ru)

---

## Клонирование репозитория

```sh
git clone https://github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests.git
cd Service-for-assigning-reviewers-for-Pull-Requests
```

## Запуск через Docker Compose

### Предварительные требования

- Docker
- Docker Compose

### Запуск

```
docker compose up -d --build
```

Compose поднимает:

- PostgreSQL (основная БД сервиса).
- Reviewer Service (наш Go-сервис).
- pgAdmin (для просмотра БД; доступен на http://localhost:5050).

После успешного запуска:

- HTTP API сервиса доступен на: http://localhost:8080
- Swagger UI — см. раздел ниже.

Все миграции из каталога migrations/ применяются автоматически при старте сервиса (через internal/repo/postgres.RunMigrations), никаких дополнительных действий руками не требуется.

## Конфигурация

Модель конфига живёт в internal/config/config.go. Основные секции:

- http
  - addr — адрес HTTP-сервера (например, ":8080").
- database — параметры подключения к PostgreSQL.

Конфиг загружается из YAML-файла с помощью функций из [internal/config/config.go](./internal/config/config.go), путь задаётся флагом -config.

## API и Swagger
### OpenAPI-спецификация

Исходная спецификация:
`api/openapi/openapi.yml`


Сгенерированный Go-код:
`api/openapi/openapi.gen.go`


Перегенерация осуществляется через oapi-codegen (конкретная команда вынесена в Makefile; там же можно посмотреть актуальный таргет для регенерации).

### Swagger UI

Swagger UI доступен по URL:
- `http://localhost:8080/swagger`

- Корневой путь / настроен на редирект на Swagger, чтобы из браузера сразу попадать в интерактивную документацию и можно было тестить ручки.
