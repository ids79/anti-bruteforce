## Общее описание
Сервис предназначен для борьбы с подбором паролей при авторизации в какой-либо системе.

Сервис вызывается перед авторизацией пользователя и может либо разрешить, либо заблокировать попытку.

Предполагается, что сервис используется только для server-server, т.е. скрыт от конечного пользователя.

## Алгоритм работы
Сервис ограничивает частоту попыток авторизации для различных комбинаций параметров
Пример настроек по умолчанию:
* не более N = 10 попыток в минуту для данного логина.
* не более M = 100 попыток в минуту для данного пароля (защита от обратного brute-force).
* не более K = 1000 попыток в минуту для данного IP (число большое, т.к. NAT).

Для подсчета и ограничения частоты запросов, используется алгоритм leaky bucket.

Сервис поддерживает множество bucket-ов, по одному на каждый логин/пароль/ip.

Bucket-ы хранятся в памяти (при этом реализован алгоритм удаление просроченных записей).

Сервис поддерживает White/black листы.
White/black листы содержат списки адресов сетей, которые обрабатываются более простым способом:
* Если входящий IP в whitelist, то сервис безусловно разрешает авторизацию (ok=true);
* Если - в blacklist, то отклоняет (ok=false).

## Архитектура
Микросервис состоит из API, базы данных для хранения настроек и black/white списков.
Опционально - хранилище для bucket'ов. Сервис должен предоставлять GRPC **или** REST API.

## Описание методов API

### Попытка авторизации
Запрос:
* login
* password
* ip

Ответ:
* ok (true/false) - сервис должен возвращать ok=true, если считает что запрос нормальный 
и ok=false, если считает что происходит bruteforce.

### Сброс bucket
Должен очистить bucket-ы соответствующие переданным login и ip.
* login
* ip

### Добавление IP в blacklist
* подсеть (IP + маска)

### Удаление IP из blacklist
* подсеть (IP + маска)

### Добавление IP в whitelist
* подсеть (IP + маска)

### Удаление IP из whitelist
* подсеть (IP + маска)

---

- Достаточно IPv4
- Пример подсети: 192.1.1.0/25 - представляет собой адрес 192.1.1.0 с маской 255.255.255.128
- Во время работы сервиса при поступлении очередного IP мы проходимся по подсетям в черных и белых списках и вычисляем, принадлежит ли IP одной из них.

---

## Конфигурация
Основные параметры конфигурации: N, M, K - лимиты по достижению которых, сервис считает попытку брутфорсом.

## Command-Line интерфейс
Необходимо так же разработать command-line интерфейс для ручного администрирования сервиса.
Через CLI должна быть возможность вызвать сброс бакета и управлять whitelist/blacklist-ом.
CLI может работать через GRPC/HTTP интерфейс.

## Развертывание
Развертывание микросервиса должно осуществляться командой `make run` (внутри `docker compose up`)
в директории с проектом.

## Тестирование
Рекомендуется выделить модуль обработки одного bucket и протестировать его с помощью unit-тестов.

Так же необходимо написать интеграционные тесты, проверяющие все вызовы API.
