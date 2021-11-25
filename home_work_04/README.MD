### Описание эндпойнтов и примеры запросов к ним:

- ###### POST /users/register - регистрация пользователя
        curl -v -X POST -H "Content-Type: application/json" --data '{"Login":"Login", "Passwd":"pass"}' 'localhost:5000/users/register'
- ###### GET /messages - запросить сообщения общего чата (Возможен вывод с пагинацией)
        curl -v -X GET -H "Authorization: Bearer {token}" 'localhost:5000/messages?from=1&to=5'
- ###### POST /messages - отправить сообщение для всех
        curl -v -X POST --data '{"Message":"text of message"}' -H "Authorization: Bearer {token}" 'localhost:5000/messages'
- ###### GET /messages/my - запросить личные сообщения
        curl -v -X GET -H "Authorization: Bearer {token}" 'localhost:5000/messages/my'
- ###### POST /messages/personal - отправить личное сообщение
        curl -v -X POST --data '{"Message":"Text of message", "To":"UserLogin"}' -H "Authorization: Bearer {token}" 'localhost:5000/messages/personal'


