# otus-microservices-homework-05

RESTful CRUD with Helm для домашней работы на OTUS

## Сборка и публикация
1. `docker build --platform linux/amd64 -t paulrozhkin/otus-microservices-homework-05:latest .` 
2. `docker push paulrozhkin/otus-microservices-homework-05:latest`
3. `docker run -p 8000:8000 paulrozhkin/otus-microservices-homework-05:latest`

## Работа с helm
1. Переходим в helm дирректорию:
```aiignore
cd ./deployment/helm/users-serivce
```
2. Добавляем репозитории (bitnami из под VPN):
```aiignore
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update
```
3. Подтягиваем зависимости и устанавливаем chart:
```aiignore
helm dependency build .

helm upgrade --install users-service . `
  -n users-service `
  --create-namespace `
  --wait
```
4. Если minikube запущен с driver=docker, то выполнить тунелирование:
```aiignore
minikube tunnel
```
5. Выполнить postman тест:
```aiignore
newman run ./../../postman/OTUS-homework-5.postman_collection.json
```