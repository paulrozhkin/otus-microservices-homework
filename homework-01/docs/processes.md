# Диаграммы процессов

[К решению](../README.md) / [Контейнеры](containers.md) / [Контракты](contracts.md)

На этих диаграммах разобраны основные сценарии из задания. Чтобы схемы не были слишком широкими, UI и Gateway указаны в подписи пользователя, а базы и брокер отдельно не показаны. Сообщения с пометкой `event:` проходят через брокер с outbox/inbox. Сплошные стрелки - запросы и действия, пунктирные - ответы и события. Стрелка к самому сервису означает локальную проверку или запись данных.

## PV-01. Расчет и создание заказа - UC-01, UC-02, UC-03

```mermaid
sequenceDiagram
    autonumber
    actor C as Покупатель через UI и Gateway
    participant O as Orders
    participant S as Stores
    participant M as Menu
    participant F as Fulfillment
    C->>O: POST /v1/checkout-quotes
    O->>S: POST /v1/stores/{storeId}/eligibility
    S-->>O: Способы получения, валюта, тариф, доступность
    O->>M: POST /v1/quotes
    M-->>O: Позиции, скидки, сумма блюд, expiresAt
    O->>O: Сохранить полный расчет с доставкой
    O-->>C: 201 quoteId, total, expiresAt
    C->>O: POST /v1/orders с quoteId и Idempotency-Key
    O->>O: Проверить срок, сохранить заказ и outbox
    O-->>C: 201 orderId, PENDING_ACCEPTANCE, accessToken
    O-->>F: event: OrderPlaced
    F->>F: Сохранить задание и inbox
```

Показан успешный расчет. Недоступная доставка или блюдо возвращают бизнес-ошибку без создания заказа; просроченный расчет требует нового подтверждения цены. Ответ `201` означает сохранение заказа, а не согласие кухни. Подтверждение точки показано в следующих процессах. Повтор команды с тем же ключом не создает второй заказ.

## PV-02. Онлайн-оплата и самовывоз - UC-02, UC-04, UC-06, UC-09

К этому моменту заказ уже создан по сценарию PV-01 и попал в очередь точки.

```mermaid
sequenceDiagram
    autonumber
    actor S as Сотрудник через UI и Gateway
    actor C as Покупатель через UI и Gateway
    participant F as Fulfillment
    participant O as Orders
    participant P as Payments
    participant X as Платежный провайдер
    participant Maps as Карты A или B
    S->>F: POST transitions, ACCEPT и readyAt
    F->>F: Сохранить резерв и событие
    F-->>O: event: FulfillmentAccepted
    O->>P: POST /v1/payments
    P->>X: Создать платеж
    X-->>P: Внешний ID и paymentUrl
    P-->>O: 201 paymentId, paymentUrl
    C->>O: GET /v1/orders/{orderId}
    O-->>C: readyAt, paymentUrl, ожидается оплата
    C->>F: POST /v1/routes, маршрут до точки
    F->>Maps: Запросить маршрут с пробками
    Maps-->>F: Маршрут и длительность
    F-->>C: directionsUrl, durationSeconds, trafficAvailable
    C->>X: Оплатить на странице провайдера
    X->>P: Подписанный webhook об успешной оплате
    P->>P: Проверить подпись, сумму, сохранить платеж и outbox
    P-->>O: event: PaymentSucceeded
    O->>O: Проверить состояние и срок, сохранить разрешения
    O-->>F: event: PreparationAuthorized
    O-->>F: event: HandoverAuthorized
    S->>F: POST transitions, START
    F-->>O: event: FulfillmentUpdated PREPARING
    S->>F: POST transitions, READY
    F-->>O: event: FulfillmentUpdated READY
    C->>O: GET /v1/orders/{orderId}
    O-->>C: Заказ готов, время и адрес точки
    S->>F: POST transitions, COMPLETE после выдачи
    F-->>O: event: FulfillmentUpdated COMPLETED
```

Здесь оплата обработана до окончания резерва. При отказе точки `FulfillmentRejected` завершает процесс без создания платежа. При ошибке карт сохраняются адрес и время готовности; оплата и приготовление продолжаются. Переход `COMPLETE` проверяет готовность и наличие `HandoverAuthorized`. Возврат браузера со страницы оплаты не используется как подтверждение списания.

## PV-03. Доставка с оплатой при получении - UC-03, UC-05, UC-06

Заказ создан с `fulfillmentType=DELIVERY` и `paymentMethod=COD`. Адрес доставки проверен при оформлении в PV-01.

```mermaid
sequenceDiagram
    autonumber
    actor S as Сотрудник через UI и Gateway
    actor D as Водитель через UI и Gateway
    participant F as Fulfillment
    participant O as Orders
    participant P as Payments
    participant Maps as Карты A или B
    S->>F: POST transitions, ACCEPT и readyAt
    F-->>O: event: FulfillmentAccepted
    O-->>F: event: PreparationAuthorized
    S->>F: POST transitions, START
    F-->>O: event: FulfillmentUpdated PREPARING
    S->>F: POST transitions, READY
    F-->>O: event: FulfillmentUpdated READY
    S->>F: POST transitions, ASSIGN_DRIVER с driverId
    D->>F: POST /v1/routes, от точки к покупателю
    F->>Maps: Получить маршрут и время с пробками
    Maps-->>F: Маршрут и оценка времени
    F-->>D: Маршрут и длительность
    D->>F: POST transitions, DISPATCH и deliveryEta
    F-->>O: event: FulfillmentUpdated DISPATCHED
    Note over D,P: Водитель прибыл и получил оплату от покупателя
    D->>O: POST /v1/orders/{orderId}/collection
    O->>P: POST /v1/payments/collections
    P->>P: Сохранить факт оплаты и outbox
    P-->>O: 201 paymentId, SUCCEEDED
    O-->>D: 202 регистрация инициирована
    P-->>O: event: PaymentSucceeded
    O-->>F: event: HandoverAuthorized
    D->>F: POST transitions, COMPLETE после выдачи
    F-->>O: event: FulfillmentUpdated COMPLETED
```

Оплата при получении не вызывает онлайн-платежного провайдера: Payments учитывает факт, зарегистрированный водителем с нужными правами. Приготовление разрешено ранее, поэтому повторное разрешение приготовления по `PaymentSucceeded` не требуется. Если событие оплаты еще не доставлено, `COMPLETE` возвращает конфликт состояния; интерфейс обновляет задание и позволяет повторить действие после разрешения. Без водителя отправка невозможна, сотрудник уточняет время или фиксирует невозможность исполнения через `FAIL`.

## PV-04. Истечение резерва и поздняя оплата - альтернатива UC-04

Точка уже приняла заказ и платеж создан, но Orders еще не получил подтверждение оплаты.

```mermaid
sequenceDiagram
    autonumber
    participant O as Orders
    participant F as Fulfillment
    participant P as Payments
    participant X as Платежный провайдер
    actor C as Покупатель через UI и Gateway
    O->>O: Дедлайн истек, атомарно установить CANCELLED и outbox
    O-->>F: event: OrderCancelled
    F->>F: Освободить резерв, сохранить отмену
    X->>P: Поздний webhook об успешном списании
    P->>P: Проверить и сохранить платеж с outbox
    P-->>O: event: PaymentSucceeded
    O->>O: Заказ отменен, установить REFUND_PENDING
    O->>P: POST /v1/payments/{paymentId}/refunds
    P->>X: Запросить полный возврат с постоянным ключом
    X-->>P: Возврат принят в обработку
    P-->>O: 202 refundId, PENDING
    X->>P: Подтверждение возврата
    P-->>O: event: PaymentRefunded
    O->>O: Сохранить REFUNDED, заказ остается CANCELLED
    C->>O: GET /v1/orders/{orderId}
    O-->>C: CANCELLED, paymentStatus=REFUNDED
```

Статус заказа и статус оплаты раздельны. До подтверждения возврата отображается `REFUND_PENDING`; при временной ошибке возврат повторяется с тем же ключом. Если повторы не помогли, ошибку разбирает сотрудник. Если оплата не поступит, возврат не создается. Невозможность исполнения уже оплаченного заказа (`FulfillmentFailed`) приводит к той же компенсации.
