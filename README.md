# Go Pet Shop

Учебный REST API интернет-магазина товаров для питомцев на Go и PostgreSQL.
Проект развивается по четырём последовательным версиям, каждая из которых
хранится в отдельной ветке.

## Версии проекта

| Ветка | Содержание |
|---|---|
| `v1` | CRUD пользователей и товаров |
| `v2` | Заказы и позиции заказа |
| `v3` | Атомарное оформление заказа |
| `v4` | История заказов и аналитика |

Полное приложение находится в ветке `v4`. Она наследует весь API `v3` и
добавляет историю заказов с оплатами, а также рейтинг проданных товаров.

### Что добавлено по версиям и зачем

- `v1`: модели, SQL и REST CRUD для пользователей и товаров — основа работы с
  простыми `INSERT`, `SELECT`, `UPDATE` и `DELETE`.
- `v2`: связанные таблицы заказов и позиций, внешние ключи и `JOIN` — практика
  связей между сущностями и получения составных данных.
- `v3`: таблица финансовых транзакций и атомарный checkout — остатки, заказ,
  позиции и оплата либо сохраняются вместе, либо вместе откатываются.
- `v4`: история через несколько `JOIN` и популярность через `SUM`/`GROUP BY` —
  практика составных аналитических SQL-запросов.

## Быстрый запуск

Нужен Docker с поддержкой Docker Compose. Приложение, PostgreSQL и миграции
запускаются одной командой:

```bash
git clone https://github.com/AlexanderVodyanitsky/go-pet-shop.git
cd go-pet-shop
git switch v4
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

## API версии v4

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
| `GET` | `/users/history?email={email}` | полная история заказов и оплат |
| `GET` | `/users/{email}/history` | альтернативный адрес истории |
| `GET` | `/products/popular` | товары по убыванию проданного количества |
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

История пользователя возвращает каждый заказ вместе с массивом позиций, суммой
и состоянием финансовой транзакции:

```bash
curl 'http://localhost:8080/users/history?email=alex@example.com'
```

Рейтинг популярности учитывает только позиции заказов с успешной транзакцией
`completed`:

```bash
curl http://localhost:8080/products/popular
```

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
internal/service/        бизнес-правила и сценарии использования
internal/storage/        общие ошибки репозиториев
internal/storage/postgres/ SQL-запросы PostgreSQL
migrations/              последовательные миграции схемы
```

## Архитектура приложения

Запрос проходит через слои в одном направлении:

```text
HTTP handler → service/usecase → repository/storage → PostgreSQL
```

- Handler отвечает только за HTTP: декодирует JSON и параметры, вызывает один
  метод service и переводит результат в HTTP-статус.
- Service нормализует и проверяет доменные данные, выполняет сценарий приложения
  и объединяет несколько операций repository.
- PostgreSQL storage содержит только SQL и преобразование ошибок базы данных.

Благодаря этому бизнес-правила тестируются без HTTP и PostgreSQL, а handlers не
зависят от конкретной реализации хранилища.
