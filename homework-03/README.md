# otus-microservices-homework-03

Простой http сервер с запуском в docker для домашней работы на OTUS

## Сборка и публикация
1. `docker build --platform linux/amd64 -t paulrozhkin/otus-microservices-homework-02:latest .` 
2. `docker push paulrozhkin/otus-microservices-homework-02:latest`
3. `docker run -p 8000:8000 paulrozhkin/otus-microservices-homework-02:latest`

## Работа с k8s
1. Создаем namespace
2. Переходим в k8s дирректорию:
`cd ./k8s`
3. Создаем namespace:
`kubectl create namespace health-service`
4. Устанавливаем nginx ingress:
```aiignore
cd ./helm
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx/
helm repo update
helm install nginx ingress-nginx/ingress-nginx --namespace health-service -f nginx-ingress.yaml
cd ../
```
5. Применяем манифесты
```aiignore
cd ./app
kubectl apply -n health-service -f .
cd ..
```
6. Если minikube запущен с driver=docker, то выполнить тунелирование:
```
`minikube tunnel
```
7. Выполнить postman тест:
```aiignore
cd ./postman
newman run OTUS-homework-3.postman_collection.json
```
