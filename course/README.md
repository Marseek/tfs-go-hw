# Курсовой проект TFS: Golang 2021

Целью курсового проекта является написание торгового робота на платформе [kraken-demo](https://futures.kraken.com/ru.html).
## Описание этапов работы программы:
* После начала работы программа ждет получения данных по параметрам сделки(order) и сигнала для начала работы через свое API. 
* Как только параметры сделки и сигнал к старту получены, робот подкулючается к WebSocket каналу(wss://demo-futures.kraken.com/ws/v1)
для отслеживания цен, и на основе анализа цены отправляет http запрос к АPI биржи на открытие сделки(/api/v3/sendorder), после этого
робот продолжает следить за ценой для принятия решения о закрытии сделки.
* Об открытой сделке делается запись в Postgres и отправляется сообщение в телеграмм.
* Сигнал о закрытии сделки робот получает из настроек stop-loss/take-profit(если цена переходит заданное значение - сделка закрывается).
* Сделку так же можно закрыть послав сигнал к закрытию через API робота.
* О закрытии сделки делается запись в Postgres и отправляется сообщение в Телеграм.
* После закрытия сделки робот снова ждет сигнала о начале работы
* Сообщения в телеграм отправляются отдельной программой, общение с которой происходит по gRPC

##  Информация по запуски и управлению роботом:

1. Сначала нужно запустить программу, которая принимает сообщения для отправки в телеграмм, 
которая находится в `gRPC_telegram/cmd/server/server.go`
2. Затем запускается контейнер с Postgres командой `docker-compose up` в корневой папке программы
3. Далее запускается основная программа из `api/main.go`
4. Ключи для работы с API kraken передаются через параметры запуска программы(argv), в формате: <br>
`-privat Privat_Key -public Public_Key`

##### После запуска программы управление роботом осущенствляется через API. Для дуступа к API необходма аутентификация при помощи jwt.

- ###### POST /login - Api endpoint для получения jwt токена.
`curl -v -X POST -H "Content-Type: application/json" --data '{"login": "jlexie", "passwd": "passwd"}' 'localhost:5000/login'`

##### Описание эндпойнтов для управления роботом и примеры запросов к ним:

- ###### POST /api/start - Отправить сигнал к началу работы.
`curl -v -X POST -H "Content-Type: application/json" 'localhost:5000/api/start'`

- ###### POST /api/stop - Отправить сигнал к остановке работы.
`curl -v -X POST -H "Content-Type: application/json" 'localhost:5000/api/stop'`

- ###### POST /api/set - Задать параметры сделки.
`curl -v -X POST -H "Content-Type: application/json" --data '{"ticker": "PI_XBTUSD", "size": 2, "profit": 0.05, "side":"buy"}' 'localhost:5000/api/set'` <br>
"ticker" - инструмент, "size" - размер сделки, "side" - направление сделки, "profit" - stop-loss/take-profit в процентах от цены

- ###### POST /api - Задать параметры сделки и отправить сигнал к старту работы.
`curl -v -X POST -H "Content-Type: application/json" --data '{"start": 1, "ticker": "PI_XBTUSD", "size": 2, "profit": 0.05, "side":"buy"}' 'localhost:5000/api/set'` <br>
"start" - может быть 1 или 0. 1 - старт, 0 - cтоп.

Каждый запрос к /api ендпойнтам должен включать в себя header c токеном jwt:
`-H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NDEzNjMzNzcsImlhdCI6MTYzNzY3Njk3NywiTG9naW4iOiJqbGV4aWUifQ.JUr3hVS4c0-HrbzKCMCJrLbAn34TVg3NKXXRdXU-e2g"
`
