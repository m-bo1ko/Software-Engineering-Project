# EMSIB - Enterprise Microservices Infrastructure for Buildings

Мікросервісна інфраструктура для управління будівлями та енергоефективністю.

## Архітектура

Проект складається з кількох мікросервісів:

- **security-service** (порт 8080) - Аутентифікація, авторизація, управління користувачами та ролями
- **forecast-service** (порт 8082) - Прогнозування енергоспоживання та оптимізація

## Спільна MongoDB

Всі сервіси використовують **одну спільну MongoDB** базу даних, але кожен сервіс має свою окрему базу даних всередині MongoDB:

- `security_service` - база даних для security-service
- `forecast_service` - база даних для forecast-service

## Швидкий старт

### Запуск всіх сервісів зі спільною MongoDB

Використовуйте головний `docker-compose.yml` в корені проекту:

```bash
# З кореня проекту
docker-compose up -d

# Перевірити статус
docker-compose ps

# Переглянути логи
docker-compose logs -f

# Зупинити всі сервіси
docker-compose down
```

Це підніме:
- MongoDB на порту 27017
- Security Service на порту 8080
- Forecast Service на порту 8082
- MongoDB Express (опціонально, профіль dev) на порту 8081

### Запуск MongoDB Express (UI для БД)

```bash
docker-compose --profile dev up -d mongo-express
```

Доступ: http://localhost:8081
- Username: admin
- Password: admin123

### Запуск окремих сервісів (standalone)

Якщо потрібно запустити окремий сервіс локально:

1. Спочатку запустіть спільну MongoDB:
```bash
# З кореня проекту
docker-compose up -d mongodb

# Або створіть network вручну
docker network create emsib-network
docker run -d --name emsib-mongodb \
  --network emsib-network \
  -p 27017:27017 \
  -v mongodb_data:/data/db \
  mongo:7.0
```

2. Потім запустіть сервіс локально:
```bash
# Security Service
cd security-service
go run cmd/main.go

# Forecast Service
cd forecast-service
go run cmd/main.go
```

## Конфігурація MongoDB

Обидва сервіси автоматично підключаються до спільної MongoDB через змінні оточення:

```bash
MONGODB_URI=mongodb://mongodb:27017  # Для Docker
MONGODB_URI=mongodb://localhost:27017  # Для локального запуску
```

Кожен сервіс використовує свою базу даних:
- `MONGODB_DATABASE=security_service` для security-service
- `MONGODB_DATABASE=forecast_service` для forecast-service

## Структура проектів

```
.
├── docker-compose.yml          # Спільний docker-compose для всіх сервісів
├── security-service/           # Security & External Integration Service
│   ├── docker-compose.yml     # Для standalone запуску (закоментована MongoDB)
│   └── ...
├── forecast-service/           # Forecast & Optimization Service
│   ├── docker-compose.yml     # Для standalone запуску (закоментована MongoDB)
│   └── ...
└── ...
```

## Перевірка роботи

### Health Checks

```bash
# Security Service
curl http://localhost:8080/health

# Forecast Service
curl http://localhost:8082/health
```

### Перевірка MongoDB

```bash
# Підключитися до MongoDB
docker exec -it emsib-mongodb mongosh

# Переглянути бази даних
show dbs

# Переглянути колекції в security_service
use security_service
show collections

# Переглянути колекції в forecast_service
use forecast_service
show collections
```

## Переваги спільної MongoDB

1. **Економія ресурсів** - одна інстанція MongoDB замість кількох
2. **Спрощене управління** - один контейнер для моніторингу
3. **Зручна розробка** - один порт (27017) для підключення
4. **Єдина точка резервного копіювання** - всі дані в одному місці
5. **Оптимізація ресурсів** - спільне використання пам'яті та процесорного часу

## Міграція з окремих MongoDB

Якщо раніше використовувалися окремі MongoDB:

1. Зупиніть старі контейнери:
```bash
cd security-service
docker-compose down

cd ../forecast-service
docker-compose down
```

2. Експортуйте дані (якщо потрібно):
```bash
# Експорт зі старої БД
docker exec security-mongodb mongodump --db=security_service --out=/backup
docker exec forecast-mongodb mongodump --db=forecast_service --out=/backup
```

3. Запустіть спільну MongoDB:
```bash
cd ..
docker-compose up -d mongodb
```

4. Імпортуйте дані:
```bash
# Імпорт в спільну БД
docker exec -i emsib-mongodb mongorestore --db=security_service /backup/security_service
docker exec -i emsib-mongodb mongorestore --db=forecast_service /backup/forecast_service
```

5. Запустіть сервіси:
```bash
docker-compose up -d
```

## Troubleshooting

### Проблема: Сервіси не можуть підключитися до MongoDB

Перевірте:
1. MongoDB запущена: `docker-compose ps mongodb`
2. Network створено: `docker network ls | grep emsib-network`
3. Змінні оточення правильні: `docker-compose config`

### Проблема: Port already in use

Якщо порт 27017 зайнятий:
```bash
# Зупинити конфліктуючий контейнер
docker ps | grep mongo
docker stop <container-id>
```

Або змінити порт в docker-compose.yml:
```yaml
ports:
  - "27018:27017"  # Замість 27017
```

## Додаткова інформація

- Security Service: див. `security-service/README.md`
- Forecast Service: див. `forecast-service/README.md`

