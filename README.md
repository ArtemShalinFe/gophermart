# Gophermart

![Made](https://img.shields.io/badge/Made%20with-Go-1f425f.svg) [![codecov](https://codecov.io/gh/ArtemShalinFe/gophermart/branch/master/graph/badge.svg?token=1H84IB1DO1)](https://codecov.io/gh/ArtemShalinFe/gophermart)

Индивидуальный дипломный проекта курса «Go-разработчик».

## Требования к окружению

- [docker](https://docs.docker.com/engine/install/)
- [docker compose](https://docs.docker.com/compose/install/linux/)
- [jq](https://jqlang.github.io/jq/download/)
- [go](https://go.dev/doc/install)

## Как запустить

### Локальный запуск

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. Из каталога `deployments` выполните команду

```sh
docker compose --env-file .env up -d --force-recreate 
```

### Запуск тестов

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. Из корневого каталога выполните команду

```sh
go test ./... -v -race
```

## Дорожная карта

- [ ] Разработка
  - [ ] Слой бизнес логики
  - [ ] Слой работы с БД
  - [ ] Слой обработки входящих запросов
  - [ ] Слой взамодействия с accrual
- [ ] Разработка юнит-тестов
- [ ] Реализовать изменение схемы БД через миграции
- [x] Подключить codecov
- [x] Добавить github badges
- [x] Добавить локальный запуск сервиса при помощи docker compose
- [ ] Сгенерировать интерфейс сервиса при помощи OpenAPI
- [x] Написать README
