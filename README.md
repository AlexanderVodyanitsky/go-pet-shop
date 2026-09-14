# Go Pet Shop

Учебный REST API интернет-магазина товаров для питомцев на Go и PostgreSQL.
Проект развивается по четырём последовательным версиям, каждая из которых
хранится в отдельной ветке.

## Версии проекта

| Ветка | Содержание | Статус |
|---|---|---|
| `v1` | CRUD пользователей и товаров | готово |
| `v2` | Заказы и позиции заказа | готово |
| `v3` | Атомарное оформление заказа | готово |
| `v4` | История заказов и аналитика | следующая версия |

Текущая ветка `v3` наследует весь API `v2` и добавляет таблицу `transactions`, а
также атомарное оформление заказа. Получение позиций заказа использует `JOIN` с
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

## API версии v3

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
| `POST` | `/checkout` | атомарно оформить и оплатить заказ |
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

Низкоуровневые операции `POST /orders` и `POST /orders/{id}/items` сохранены для
демонстрации простых `INSERT` из `v2`. Для обычного оформления используется
`POST /checkout`:

```bash
curl -i -X POST http://localhost:8080/checkout \
  -H 'Content-Type: application/json' \
  -d '{
    "user_email":"alex@example.com",
    "items":[
      {"product_id":1,"quantity":2}
    ]
  }'
```

Checkout выполняет в одной транзакции PostgreSQL:

1. Проверку пользователя.
2. Уменьшение `stock` только при достаточном остатке.
3. Расчёт полной суммы по ценам из базы.
4. Создание заказа и его позиций.
5. Создание финансовой транзакции со статусом `completed`.

Ошибка на любом шаге приводит к `ROLLBACK`. Повторяющиеся товары во входном
массиве объединяются, а блокировка товаров идёт в стабильном порядке, что
снижает риск взаимных блокировок при одновременных покупках.

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
