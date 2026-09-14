# Go Pet Shop

Учебный REST API интернет-магазина товаров для питомцев на Go и PostgreSQL.
Проект развивается по четырём последовательным версиям, каждая из которых
хранится в отдельной ветке.

## Версии проекта

| Ветка | Содержание | Статус |
|---|---|---|
| `v1` | CRUD пользователей и товаров | готово |
| `v2` | Заказы и позиции заказа | готово |
| `v3` | Атомарное оформление заказа | следующая версия |
| `v4` | История заказов и аналитика | следующая версия |

Текущая ветка `v2` наследует весь API `v1` и добавляет таблицы `orders` и
`order_items` отдельной миграцией. Получение позиций заказа использует `JOIN` с
таблицей товаров и возвращает название и текущую цену товара.

## Быстрый запуск

Нужен Docker с поддержкой Docker Compose. Приложение, PostgreSQL и миграции
запускаются одной командой:

```bash
docker compose up --build
```

API будет доступно по адресу `http://localhost:8080`. Проверка состояния:

```bash
curl http://localhost:8080/status
```

Настройки по умолчанию уже подходят для Docker. Чтобы изменить логин, пароль,
имя базы или таймаут, скопируйте `.env.example` в `.env` и отредактируйте значения.

## Локальный запуск без контейнера приложения

1. Создайте `.env` из примера и укажите доступный PostgreSQL.
2. Примените миграции: `task migrate`.
3. Запустите API: `task run`.

## API версии v2

| Метод | URL | Результат |
|---|---|---|
| `POST` | `/users` | создать пользователя |
| `GET` | `/users` | получить всех пользователей |
| `GET` | `/users/{email}` | получить пользователя по email |
| `POST` | `/products` | создать товар |
| `GET` | `/products` | получить все товары |
| `GET` | `/products/{id}` | получить товар по ID |
| `PUT` | `/products/{id}` | полностью обновить товар |
| `DELETE` | `/products/{id}` | удалить товар |
| `POST` | `/orders` | создать заказ |
| `POST` | `/orders/{id}/items` | добавить позицию в заказ |
| `GET` | `/orders/{id}` | получить заказ вместе с позициями |
| `GET` | `/users/orders?email={email}` | получить заказы пользователя |
| `GET` | `/users/{email}/orders` | альтернативный адрес заказов пользователя |
| `GET` | `/status` | проверить состояние API |

Пример создания пользователя:

```bash
curl -i -X POST http://localhost:8080/users \
  -H 'Content-Type: application/json' \
  -d '{"name":"Alex","email":"alex@example.com"}'
```

Пример создания товара:

```bash
curl -i -X POST http://localhost:8080/products \
  -H 'Content-Type: application/json' \
  -d '{"name":"Корм для кошек","price":12.50,"stock":20}'
```

Пример создания заказа и добавления позиции:

```bash
curl -i -X POST http://localhost:8080/orders \
  -H 'Content-Type: application/json' \
  -d '{"user_email":"alex@example.com","total_price":25.00}'

curl -i -X POST http://localhost:8080/orders/1/items \
  -H 'Content-Type: application/json' \
  -d '{"product_id":1,"quantity":2}'
```

В `v2` эти две операции намеренно независимы: версия показывает простые
`INSERT` и `JOIN`. Атомарное создание заказа, расчёт суммы и уменьшение остатков
реализуются следующим шагом в `v3` через транзакцию.

При успешном создании API возвращает `201 Created`. Некорректные данные дают
`400 Bad Request`, повторный email — `409 Conflict`, отсутствующая запись —
`404 Not Found`, а непредвиденная внутренняя ошибка — `500 Internal Server Error`.

## Проверки качества

```bash
task test
task linter
```

Если Task не установлен, тесты можно запустить напрямую: `go test ./...`.

## Структура

```text
cmd/app/                 запуск HTTP-сервера
cmd/migrator/            запуск SQL-миграций
config/                  конфигурация окружений
internal/handlers/       HTTP-обработчики
internal/models/         модели API и базы данных
internal/storage/        ошибки и контракты хранилища
internal/storage/postgres/ SQL-запросы PostgreSQL
migrations/              последовательные миграции схемы
```
