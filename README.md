# Gaus Storage: S3-like Microservices Cloud Storage

## Описание

Gaus Storage — это учебный проект облачного хранилища файлов, реализованный на Go в виде набора микросервисов. Архитектура построена вокруг трех основных сервисов:

- **Auth Service** — сервис аутентификации и управления пользователями (порт 8000)
- **S3 Service** — сервис хранения файлов (порт 8080)
- **API Gateway** — фронтовой шлюз, объединяющий доступ к сервисам через единый API и web-интерфейс (порт 7000)

В качестве базы данных используется PostgreSQL. Все сервисы конфигурируются через TOML/YAML файлы.

## Архитектура

```
┌─────────────┐         ┌─────────────┐         ┌─────────────┐
│  Frontend   │◀──────▶│ API Gateway │◀──────▶│  S3 Service │
└─────────────┘         └──────┬──────┘         └──────┬──────┘
                               │                       │
                               ▼                       ▼
                          Auth Service             PostgreSQL
```

- **API Gateway** принимает все HTTP-запросы клиентов, маршрутизирует их к нужным микросервисам, занимается CORS, логированием, проксированием и обеспечивает единый вход для фронта и API.
- **Auth Service** реализует регистрацию, вход, выход, управление пользователями и JWT-аутентификацию.
- **S3 Service** реализует загрузку, скачивание, удаление, шаринг и хранение файлов. Хранит файлы на диске и метаданные в базе данных.

## Основные возможности

- Регистрация и аутентификация пользователей через `Auth Service`
- JWT/Cookie-авторизация между сервисами через API Gateway
- Загрузка, скачивание, удаление, публикация (шаринг) файлов через API
- Файлы хранятся на диске, метаданные — в PostgreSQL
- REST API (см. ниже)
- Минималистичный фронтенд (HTML+JS), работающий через API Gateway

## Быстрый старт

### Зависимости

- Go 1.21+
- PostgreSQL
- make (опционально для сборки и запуска)

### Запуск

1. **Склонируйте репозиторий:**
    ```sh
    git clone https://github.com/Manabreaker/S3_project.git
    cd S3_project
    ```

2. **Поднимите базы данных и выполните миграции:**
    - Создайте БД `users` и `s3` в PostgreSQL
    - Выполните SQL-миграции из папок `auth/migrations` и `S3/migrations`

3. **Соберите и запустите сервисы:**
    ```sh
    make build
    make run
    ```
    Или вручную:
    ```sh
    go run auth/cmd/auth/main.go
    go run S3/cmd/S3/main.go
    go run APIGateway/cmd/gateway/main.go
    ```

4. **Откройте браузер:**
    - Основной интерфейс: [http://localhost:7000](http://localhost:7000)
    - API доступен по тем же адресам (см. ниже)

## API

### Auth Service (через API Gateway)

- `POST /register` — регистрация пользователя
- `POST /login` — вход (авторизация)
- `POST /logout` — выход

### S3 Service (через API Gateway)

- `GET /files` — получить список файлов пользователя
- `POST /upload` — загрузить файл
- `POST /download` — скачать файл
- `DELETE /delete` — удалить файл
- `POST /share` — создать публичную ссылку
- `GET /share/{uuid}` — страница публичного файла (фронт)
- `GET /file/{uuid}` — получить содержимое публичного файла

Все запросы кроме `/register` и `/login` требуют авторизации (cookie с JWT).

### Пример запроса загрузки файла

```http
POST /upload
Content-Type: application/json
Cookie: Authorization=...

{
  "filename": "myphoto.jpg",
  "file": "<base64>"
}
```

## Конфигурация

- Файлы конфигурации сервисов:  
  - `auth/configs/apiserver.toml`
  - `S3/configs/apiserver.toml`
  - `APIGateway/configs/apiserver.yml`

## Тесты

- Юнит- и интеграционные тесты:  
  ```sh
  make test
  ```

## Ссылки

- [Документация по API](https://github.com/Manabreaker/go_s3/blob/main/API_Docs)
- [Примеры миграций](auth/migrations/ и S3/migrations/)
- [Makefile](Makefile)

## Примечания

- В проекте реализован базовый web-интерфейс (HTML+JS) для работы с файлами через API Gateway.  
- Вся бизнес-логика реализована на Go, фронт — минимален и служит для демонстрации.

---

**Автор:** Manabreaker  
**Год:** 2025
