Учебный проект на Golang.  
Для подключения к MongoDB через используется подключение `mongodb://admin:password@localhost:27017/logs?authSource=admin&readPreference=primary&appname=MongDB%20Compass&ssl=false`.  
Для запуска миграций в консоли перейти в `auth-service/cmd/api` и прописать `migrate -path db/migrations -database "postgres://postgres:password@localhost:5432/test_task?sslmode=disable" up`.  
Для запуска проекта в Docker'e перейти в консоли в папку `project` и прописать `make up_build`. Для остановки `make down`, для запуска front-end части `make start`, для её остановки `make stop`.  
Разбит на несколько сервисов - auth-service где происходит работа с пользователями в базе данных PostgreSQL: на данном этапе регистрация и аутентификация. При аутентификации автоматически генерируется и отправляется JWT access token, который для примера используется для получения информации о всех пользователях с адреса `http://localhost:8081/users`.  
log-service - логирует происходящие события, например, аутентификацию пользователя или его регистрацию в MongoDB. Также на фронте можно отправить тестовый запрос к этому сервису для проверки его работы, может логировать через gRPC, для этого добавлена отдельная кнопка.  
broker-service - брокер, через который проходят все запросы от сервисов и к сервисам.  
Пример регистрации пользователя:  
![registr_log_postman](https://github.com/user-attachments/assets/f62e1cd2-0c66-4fe8-bd96-a5deaaf6c32a)  
Невозможность зарегестрироваться не с уникальной почты:  
![restr_of_nonunique_registr](https://github.com/user-attachments/assets/a065275e-aca6-45d0-b363-0fdc8ccf1f4f)  
Лог регистрации пользователя:  
![registr_log](https://github.com/user-attachments/assets/107d45b2-63b7-4532-a5b7-6e1152ca87b9)  
Аутентификация через в обход брокера:  
![auth_through_auth-service](https://github.com/user-attachments/assets/48884c02-ed6a-4c83-b24f-dee39ac8f3cd)  
Аутентификация через брокера:  
![auth_through_broker-service](https://github.com/user-attachments/assets/5bf84988-563f-4a58-af21-b44c34c1d49e)  
Отправка лога через gRPC:  
![grpc_log_test](https://github.com/user-attachments/assets/96a9dfbe-d911-439b-b88b-88e8f836701c)  
![grpc_log_test_successfull](https://github.com/user-attachments/assets/fef6e768-6a57-42aa-9e55-446f166d33cd)  
Тестирование лога через фронт, логи этого можно наблюдать на скринах с MongoDB с подписями "some kind of data":  
![logs_test](https://github.com/user-attachments/assets/b05148b6-249e-4bf7-8cb4-9e46f289ffd4)  

