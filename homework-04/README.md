# otus-microservices-homework-04

RESTful CRUD with Helm для домашней работы на OTUS

## Сборка и публикация
1. `docker build --platform linux/amd64 -t paulrozhkin/otus-microservices-homework-04:latest .` 
2. `docker push paulrozhkin/otus-microservices-homework-04:latest`
3. `docker run -p 8000:8000 paulrozhkin/otus-microservices-homework-04:latest`


## Работа с k8s
1. Переходим в k8s дирректорию:
```aiignore
cd ./deployment/k8s
```
2. Создаем namespace:
```aiignore
kubectl create namespace users-service
 ```
3. Устанавливаем nginx ingress:
```aiignore
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx/
helm repo update
helm install nginx ingress-nginx/ingress-nginx --namespace users-service -f nginx-ingress.yaml
```
4. Применяем манифесты
```aiignore
kubectl apply -n users-service -f .
```
5. Если minikube запущен с driver=docker, то выполнить тунелирование:
```aiignore
minikube tunnel
```
6. Выполнить postman тест:
```aiignore
newman run ./../../postman/OTUS-homework-4.postman_collection.json
```

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
newman run ./../../postman/OTUS-homework-4.postman_collection.json
```