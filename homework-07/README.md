# otus-microservices-homework-07

## Задание

Реализовать сервис заказа. Сервис биллинга. Сервис нотификаций.

При создании пользователя необходимо создавать аккаунт в сервисе биллинга. В сервисе биллинга должна быть возможность положить деньги на аккаунт и снять деньги.

Сервис нотификаций позволяет отправить сообщение на email. И позволяет получить список сообщений по методу API.

Пользователь может создать заказ. У заказа есть параметр — цена заказа.

Заказ происходит в 2 этапа:

1. Сначала снимаем деньги с пользователя с помощью сервиса биллинга.
2. Отправляем пользователю сообщение на почту с результатами оформления заказа:

   * если биллинг подтвердил платеж, должно отправиться письмо счастья;
   * если нет — письмо горя.

Упрощаем и считаем, что ничего плохого с сервисами происходить не может (они не могут падать и т. д.).

Сервис нотификаций на самом деле не отправляет сообщения, а просто сохраняет их в БД.

---

### ТЕОРЕТИЧЕСКАЯ ЧАСТЬ

#### 0. Спроектировать взаимодействие сервисов при создании заказов

Предоставить варианты взаимодействий в следующих стилях в виде **sequence-диаграммы** с описанием API на **IDL**:

* только HTTP-взаимодействие;
* событийное взаимодействие с использованием брокера сообщений для нотификаций (уведомлений);
* Event Collaboration стиль взаимодействия с использованием брокера сообщений;
* вариант, который вам кажется наиболее адекватным для решения данной задачи.

Если он совпадает с одним из вариантов выше — просто отметить это.

---

### ПРАКТИЧЕСКАЯ ЧАСТЬ

Выбрать один из вариантов и реализовать его.

На выходе должны быть:

#### I) Описание архитектурного решения

Описание архитектурного решения и схема взаимодействия сервисов **в виде картинки**.

#### II) Команда установки приложения

Команда установки приложения:

* из Helm;
* или из манифестов.

Обязательно указать, в каком **namespace** нужно устанавливать приложение.

#### III) Тесты Postman

Тесты Postman должны прогонять следующий сценарий:

1. Создать пользователя. Должен создаться аккаунт в биллинге.
2. Положить деньги на счет пользователя через сервис биллинга.
3. Сделать заказ, на который хватает денег.
4. Посмотреть деньги на счету пользователя и убедиться, что их сняли.
5. Посмотреть в сервисе нотификаций отправленные сообщения и убедиться, что сообщение отправилось.
6. Сделать заказ, на который не хватает денег.
7. Посмотреть деньги на счету пользователя и убедиться, что их количество не поменялось.
8. Посмотреть в сервисе нотификаций отправленные сообщения и убедиться, что сообщение отправилось.

#### В тестах обязательно

* наличие `{{baseUrl}}` для URL;
* использование домена `arch.homework` в качестве initial-значения `{{baseUrl}}`;
* отображение данных запроса и данных ответа при запуске из командной строки с помощью Newman.


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
go test ./services/user-service/... ./services/billing-service/... ./services/order-service/... ./services/notification-service/... ./internal/platform/...
```

Build:
```powershell
docker build --platform linux/amd64 -f services/user-service/Dockerfile -t paulrozhkin/otus-microservices-homework-07-user-service:latest .
docker build --platform linux/amd64 -f services/billing-service/Dockerfile -t paulrozhkin/otus-microservices-homework-07-billing-service:latest .
docker build --platform linux/amd64 -f services/order-service/Dockerfile -t paulrozhkin/otus-microservices-homework-07-order-service:latest .
docker build --platform linux/amd64 -f services/notification-service/Dockerfile -t paulrozhkin/otus-microservices-homework-07-notification-service:latest .
```

Push:
```powershell
docker push paulrozhkin/otus-microservices-homework-07-user-service:latest
docker push paulrozhkin/otus-microservices-homework-07-billing-service:latest
docker push paulrozhkin/otus-microservices-homework-07-order-service:latest
docker push paulrozhkin/otus-microservices-homework-07-notification-service:latest
```


## Minikube 
1. Запуска minikube с заданными ресурсами:
```
minikube start --cpus=6 --memory=8192
```

## Работа с helm
1. Переходим в helm дирректорию:
```aiignore
cd ./deployment/helm/
```
2. Добавляем репозитории:
```aiignore
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
  --set grafana.sidecar.dashboards.enabled=true `
  --set grafana.sidecar.dashboards.label=grafana_dashboard `
  --set grafana.sidecar.dashboards.searchNamespace=monitoring `
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
  
helm upgrade --install online-store . `
  -n online-store `
  --create-namespace `
  --wait `
  --wait-for-jobs
```

Проверка ресурсов мониторинга:
```aiignore
kubectl -n online-store get servicemonitors
kubectl -n monitoring get configmap -l grafana_dashboard=1
kubectl -n monitoring get configmap -l grafana_alert=1
```
6. Если minikube запущен с driver=docker, то выполнить тунелирование:
```aiignore
minikube tunnel
```
7. Выполнить postman тест:
```aiignore
newman run ./../../../tests/OTUS-homework-7.postman_collection.json --reporters cli --verbose
```

## Работа с docker compose для разработки

1. Переходим в docker compose дирректорию:
```aiignore
cd ./deployment/docker-compose
```

2. Запускаем БД и kafka
```aiignore
docker compose `
  -f deployment/docker-compose/docker-compose.yaml `
  up -d `
  kafka-init `
  kafka-ui `
  user-postgres `
  billing-postgres `
  order-postgres `
  notification-postgres
```
3. Выполняем миграции данных:
```aiigonore
go run ./services/user-service/cmd/api migrate
go run ./services/billing-service/cmd/api migrate
go run ./services/order-service/cmd/api migrate
go run ./services/notification-service/cmd/api migrate
```
4. Запускаем микросервисы в любом удобном режиме. Рекомендуется Compound run configuration in GoLand IDE
