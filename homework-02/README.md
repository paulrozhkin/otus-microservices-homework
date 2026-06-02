# otus-microservices-homework-02

Простой http сервер с запуском в docker для домашней работы на OTUS

## Сборка и публикация
1. `docker build --platform linux/amd64 -t paulrozhkin/otus-microservices-homework-02:latest .` 
2. `docker push paulrozhkin/otus-microservices-homework-02:latest`
3. `docker run -p 8000:8000 paulrozhkin/otus-microservices-homework-02:latest`