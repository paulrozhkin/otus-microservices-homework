# otus-microservices-homework-04

RESTful CRUD with Helm для домашней работы на OTUS

## Сборка и публикация
1. `docker build --platform linux/amd64 -t paulrozhkin/otus-microservices-homework-04:latest .` 
2. `docker push paulrozhkin/otus-microservices-homework-04:latest`
3. `docker run -p 8000:8000 paulrozhkin/otus-microservices-homework-04:latest`

