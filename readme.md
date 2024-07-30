Микросервис обработки сообщений
==============================
Этот микросервис предназначен для обработки сообщений, которые поступают через HTTP API.
Сообщения сохраняются в базе данных PostgreSQL, а затем отправляются в Kafka для дальнейшей обработки.
Обработанные сообщения помечаются в базе данных. 
Сервис также предоставляет API для получения статистики по обработанным сообщениям.

## Запуск
Для запуска проекта необходимо выполнить следующие команды:
```bash
docker-compose up -d
```

## Swagger
После запуска проекта, документация по API ( [docs/swagger.json](docs/swagger.json) / [docs/swagger.yaml](docs/swagger.yaml) ) будет доступна по адресу:
```
GET http://localhost:8080/api/swagger/index.html
```

## API
### Отправка сообщения
```http
POST http://localhost:8080/api/messages
```

### Получение сообщения
```http
GET http://localhost:8080/api/messages/{id}
```

### Получение сообщений
```http
GET http://localhost:8080/api/messages
```

### Получение статистики
Для статистики добавлены Prometheus и Grafana.
Результаты можно посмотреть по адресу:
```
GET http://localhost:3000
```
Для входа используйте логин и пароль: admin/admin

Prometheus доступен по адресу:
```
GET http://localhost:9090
```

Метрики доступны по адресу:
```
GET http://localhost:8080/metrics
```

Для создания дашборда с метриками, необходимо импортировать файл [grafana-dashboard-model.json](grafana-dashboard-model.json) в Grafana.
Заменив во всех `"datasource": {
"type": "prometheus",
"uid": "adtb2yxul83cwd"
},` `adtb2yxul83cwd` на ваш `uid` из вашего Prometheus.