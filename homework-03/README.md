# otus-microservices-homework-03

Простой http сервер с запуском в docker для домашней работы на OTUS

## Сборка и публикация
1. `docker build --platform linux/amd64 -t paulrozhkin/otus-microservices-homework-03:latest .` 
2. `docker push paulrozhkin/otus-microservices-homework-03:latest`
3. `docker run -p 8000:8000 paulrozhkin/otus-microservices-homework-03:latest`

## Работа с k8s
1. Переходим в k8s дирректорию:
```
1. cd ./k8s
```
2. Создаем namespace:
```
3. kubectl create namespace health-service
 ```
3. Устанавливаем nginx ingress:
```aiignore
cd ./helm
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx/
helm repo update
helm install nginx ingress-nginx/ingress-nginx --namespace health-service -f nginx-ingress.yaml
cd ../
```
4. Применяем манифесты
```aiignore
cd ./app
kubectl apply -n health-service -f .
```
5. Если minikube запущен с driver=docker, то выполнить тунелирование:
```
`minikube tunnel
```
6. Выполнить postman тест:
```aiignore
newman run ./../../postman/OTUS-homework-3.postman_collection.json
```
