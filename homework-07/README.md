# otus-microservices-homework-06

## Задание

Добавить в приложение аутентификацию и регистрацию пользователей.

Реализовать сценарий "Изменение и просмотр данных в профиле клиента".

Пользователь регистрируется, входит под своей учетной записью и по определенному URL получает данные своего профиля. Пользователь может изменить данные в профиле.

Данные профиля для чтения и редактирования не должны быть доступны другим клиентам. Это относится как к неаутентифицированным пользователям, так и к другим аутентифицированным пользователям.

На выходе должны быть:

1. Описание архитектурного решения и схема взаимодействия сервисов в виде картинки.
2. Команда установки приложения из Helm или из манифестов. Нужно обязательно указать namespace, в который устанавливается приложение.
3. Команда установки API Gateway, если используется не `nginx-ingress`.
4. Тесты Postman, которые прогоняют сценарий:
   - регистрация пользователя 1;
   - проверка, что изменение и получение профиля пользователя недоступно без логина;
   - вход пользователя 1;
   - изменение профиля пользователя 1;
   - проверка, что профиль поменялся;
   - выход, если он есть;
   - регистрация пользователя 2;
   - вход пользователя 2;
   - проверка, что пользователь 2 не имеет доступа на чтение и редактирование профиля пользователя 1.

В тестах обязательно:

1. Наличие `{{baseUrl}}` для URL.
2. Использование домена `arch.homework` в качестве initial value для `{{baseUrl}}`.
3. Использование сгенерированных случайно данных в сценарии.
4. Отображение данных запроса и данных ответа при запуске из командной строки с помощью Newman.

## Architecture

Приложение оставлено в виде одного сервиса `users-service`. Данные пользователей и сессии хранятся в PostgreSQL. После входа пользователь получает cookie `session_id`, а сама сессия сохраняется в базе. Благодаря этому проверка авторизации работает одинаково при нескольких репликах приложения.

Для входящего трафика используются два Ingress правила на одном домене `arch.homework`.

`users-public-ingress` открыт без авторизации. Через него доступны `/register`, `/login`, `/logout`, `/auth`, `/health`, `/swagger` и `/swagger.yaml`.

`users-protected-ingress` закрывает все пути `/api/v1/*`. Для этих запросов nginx-ingress сначала вызывает `/auth` через аннотацию `nginx.ingress.kubernetes.io/auth-url`.

Если сессия валидна, `/auth` возвращает заголовки `X-Auth-UserId`, `X-Auth-User`, `X-Auth-Email` и `X-Auth-Roles`. nginx-ingress добавляет эти заголовки в исходный запрос и передает его в приложение.

Методы профиля используют `X-Auth-UserId`, поэтому пользователь может читать и менять только свой профиль. Методы `/api/v1/users` дополнительно проверяют роль. Пользователь с ролью `admin` имеет полный доступ, а обычный `user` может читать и менять только запись со своим id.

![Architecture](./docs/architecture.svg)

Mermaid source: [architecture.mmd](./docs/architecture.mmd)


## Сборка и публикация
Каждый сервис со своим Go module and Dockerfile

Test:
```powershell
go test ./services/user-service/... ./services/billing-service/... ./internal/platform/...
```

Build:
```powershell
docker build --platform linux/amd64 -f services/user-service/Dockerfile -t paulrozhkin/otus-microservices-homework-07-user-service:latest .
docker build --platform linux/amd64 -f services/billing-service/Dockerfile -t paulrozhkin/otus-microservices-homework-07-billing-service:latest .
```

Push:
```powershell
docker push paulrozhkin/otus-microservices-homework-07-user-service:latest
docker push paulrozhkin/otus-microservices-homework-07-billing-service:latest
```


## Minikube 
1. Запуска minikube с заданными ресурсами:
```
minikube start --cpus=6 --memory=8192
```

## Работа с helm
1. Переходим в helm дирректорию:
```aiignore
cd ./deployment/helm/users-service
```
2. Добавляем репозитории (bitnami из под VPN):
```aiignore
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo add ingress-nginx https://kubernetes.github.io/ingress-
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
```
3. Устанавливаем мониторинг:
```aiignore
helm upgrade --install kube-prometheus-stack prometheus-community/kube-prometheus-stack `
  -n monitoring `
  --create-namespace `
  --set prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false `
  --set prometheus.prometheusSpec.podMonitorSelectorNilUsesHelmValues=false `
  --set grafana.adminPassword=admin `
  --set grafana.sidecar.alerts.enabled=true `
  --set grafana.sidecar.alerts.label=grafana_alert `
  --set grafana.sidecar.alerts.searchNamespace=monitoring `
  --set prometheus.prometheusSpec.retention=2h `
  --set prometheus.prometheusSpec.resources.requests.memory=512Mi `
  --set prometheus.prometheusSpec.resources.limits.memory=1Gi `
  --set grafana.resources.requests.memory=128Mi `
  --set grafana.resources.limits.memory=512Mi
```
4. Для доступа к стеку мониторинга следовать инструкциями из возврата предыдущей команды. 
    Нужно прокинуть доступ до Grafana через инструкции "Access Grafana local instance". Дубликат возврата:
```aiignore
Get Grafana 'admin' user password by running:

  kubectl --namespace monitoring get secrets kube-prometheus-stack-grafana -o jsonpath="{.data.admin-password}" | base64 -d ; echo

Access Grafana local instance:

  export POD_NAME=$(kubectl --namespace monitoring get pod -l "app.kubernetes.io/name=grafana,app.kubernetes.io/instance=kube-prometheus-stack" -oname)
  kubectl --namespace monitoring port-forward $POD_NAME 3000

Get your grafana admin user password by running:

  kubectl get secret --namespace monitoring -l app.kubernetes.io/component=admin-secret -o jsonpath="{.items[0].data.admin-password}" | base64 --decode ; echo


Visit https://github.com/prometheus-operator/kube-prometheus for instructions on how to create & configure Alertmanager and Prometheus instances using the Operator.
```
5. Подтягиваем зависимости и устанавливаем chart:
```aiignore
helm dependency build .

helm upgrade --install users-service . `
  -n users-service `
  --create-namespace `
  --wait
```
6. Если minikube запущен с driver=docker, то выполнить тунелирование:
```aiignore
minikube tunnel
```
7. Выполнить postman тест:
```aiignore
newman run ./../../../tests/OTUS-homework-6.postman_collection.json --reporters cli --verbose
```
8. Запустить нагрузочный тест:
```aiignore
k6 run ./../../../tests/load-test.js
```
