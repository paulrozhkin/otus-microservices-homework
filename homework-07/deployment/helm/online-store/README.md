# Online Store umbrella Helm chart

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
