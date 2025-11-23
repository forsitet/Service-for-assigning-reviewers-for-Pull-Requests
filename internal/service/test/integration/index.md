## Интеграционное тестирование

Интеграционный тест, который проверяет бизнес-логику создания и мержа Pull Request’ов на реальной PostgreSQL

### Что проверяет сценарий TestPCreateAndMerge

Тест выполняет полный сценарий:

1. Открывает реальную БД по TEST_DATABASE_CONN

2. Прогоняет миграции через postgres.RunMigrations

3. Очищает таблицы:

- teams
- users
- pull_requests
- pull_request_reviewers

4. Создаёт PRService на реальных репозиториях:

- postgres.NewUserRepo(db)

- postgres.NewPRRepo(db)

Если переменная окружения TEST_DATABASE_CONN не задана, тест помечается как skip, чтобы go test ./... можно было запускать даже без поднятой БД

### Настройка окружения для интеграционного теста

1. Поднять тестовую PostgreSQL командой `docker-compose up -d --build`

2. Указать `TEST_DATABASE_CONN` для теста

   Интеграционный тест читает строку подключения из переменной окружения `TEST_DATABASE_CONN`

   ```bash
   export TEST_DATABASE_CONN="postgres://postgres:12345@localhost:5555/reviewer-db_test?sslmode=disable"
   ```

### Как запускать интеграционный тест

Запуск только интеграционного теста PRService

```bash
go test ./internal/service/test/integration -run TestPCreateAndMerge -v
```

Если TEST_DATABASE_CONN задан и БД доступна - тест выполнится полностью

Если TEST_DATABASE_CONN не задан - тест будет пропущен
