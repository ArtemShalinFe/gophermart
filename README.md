# Gophermart

![Made](https://img.shields.io/badge/Made%20with-Go-1f425f.svg) [![codecov](https://codecov.io/gh/ArtemShalinFe/gophermart/branch/master/graph/badge.svg?token=1H84IB1DO1)](https://codecov.io/gh/ArtemShalinFe/gophermart)

Учебный проект для индивидуального дипломного проекта курса «Go-разработчик»

# Начало работы

TODO+

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без
   префикса `https://`) для создания модуля

# Обновление шаблона

TODO+

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m master template https://github.com/yandex-praktikum/go-musthave-diploma-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/master .github
```

Затем добавьте полученные изменения в свой репозиторий.
