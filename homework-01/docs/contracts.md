# Контракты взаимодействия

[К решению](../README.md) / [Контейнеры](containers.md) / [Примеры процессов](processes.md)

Ниже описаны API и события, которые нужны для выбранных сценариев. В таблицах указаны основные поля запросов и ответов. Полная OpenAPI-спецификация на этом этапе не составлялась.

## 1. Общие правила

- Синхронные запросы: HTTPS, JSON, префикс `/v1`. Идентификаторы - UUID, время - ISO 8601 UTC, деньги - `{amountMinor: integer, currency: string}`.
- Команды создания используют `Idempotency-Key`. Повтор ключа с тем же телом возвращает прежний результат, с другим телом - `409`. Ключ сохраняется в БД вместе с результатом.
- Ошибки: `{code, message, correlationId}`. HTTP `400` - неверный формат, `401/403` - доступ, `404` - объект отсутствует, `409` - конфликт состояния/цены, `422` - бизнес-ограничение, `503` - временно недоступная зависимость.
- Сервисы проверяют идентичность вызывающего и права. `storeId` из запроса не доказывает право доступа. Персонал ограничен своими точками, менеджер сети - разрешенными странами. Покупатель читает заказ по защищенному непрогнозируемому токену.
- Цена и разрешенные способы оплаты рассчитываются сервером. Клиент не задает итоговую сумму платежа или платежный аккаунт.
- Для чтений и идемпотентных команд допустимы ограниченные повторы с задержкой. Таймаут команды означает неизвестный результат: сначала повтор с тем же ключом или проверка состояния, а не создание нового объекта.

## 2. Синхронные API

| Вызывающий → сервис                            | Метод и путь                                  | Запрос                                                                                   | Успешный результат / ограничения                                                                                                                  |
|------------------------------------------------|-----------------------------------------------|------------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------|
| UI через Gateway → Stores                      | `GET /v1/stores?country=...`                  | Страна, необязательные координаты                                                        | `200`: точки, адреса, часы, timezone, способы получения и оплаты                                                                                  |
| Orders → Stores                                | `POST /v1/stores/{storeId}/eligibility`       | `fulfillmentType, deliveryAddress?`                                                      | `200`: `eligible, reason?, country, timezone, currency, paymentMethods, deliveryFee, configVersion`; самовывоз имеет нулевой тариф                |
| UI через Gateway → Menu                        | `GET /v1/stores/{storeId}/menu`               | Точка                                                                                    | `200`: блюда, доступность, цены, действующие акции                                                                                                |
| Orders → Menu                                  | `POST /v1/quotes`                             | `storeId, country, timezone, items[{productId, quantity}]`                               | `201`: `menuQuoteId, lines, discountBreakdown, subtotal, expiresAt, menuVersion`; недоступное блюдо - `422`                                       |
| UI через Gateway → Orders                      | `POST /v1/checkout-quotes`                    | `storeId, items, fulfillmentType, deliveryAddress?, paymentMethod`                       | `201`: `quoteId, lines, deliveryFee, total, expiresAt`; Orders сохраняет полный расчет на основе Stores и Menu                                    |
| UI через Gateway → Orders                      | `POST /v1/orders`                             | `quoteId, customerContact, source=WEB`                                                   | `201`: `orderId, status=PENDING_ACCEPTANCE, accessToken`; просроченный расчет - `409`, клиент запрашивает новый и подтверждает измененную цену    |
| Сотрудник через Gateway → Orders               | `POST /v1/orders`                             | Те же поля, `source=FAX`                                                                 | `201`: заказ той же очереди; право создавать `FAX` есть только у сотрудника точки                                                                 |
| UI через Gateway → Orders                      | `GET /v1/orders/{orderId}`                    | Токен доступа покупателя или токен персонала                                             | `200`: состав, итоговая цена, статус, `readyAt?, deliveryEta?, paymentStatus, paymentUrl?`                                                        |
| Orders → Payments                              | `POST /v1/payments`                           | `orderId, storeId, total, method=ONLINE, expiresAt`                                      | `201`: `paymentId, status=PENDING, paymentUrl`; один активный платеж для заказа, срок не позже резерва                                            |
| Сотрудник через Gateway → Orders               | `POST /v1/orders/{orderId}/collection`        | `method=IN_STORE/COD, receiptReference`                                                  | `202`: запрос на регистрацию оплаты принят; Orders проверяет состояние и вызывает Payments                                                        |
| Orders → Payments                              | `POST /v1/payments/collections`               | `orderId, storeId, total, method, receiptReference, actorId`                             | `201`: `paymentId, status=SUCCEEDED`; результат также публикуется событием, повтор не создает вторую оплату                                       |
| Orders → Payments                              | `POST /v1/payments/{paymentId}/refunds`       | `orderId, reason`                                                                        | `202`: `refundId, status=PENDING`; полный возврат исходного платежа, сумма берется из Payments                                                    |
| Сотрудник через Gateway → Fulfillment          | `GET /v1/tasks?storeId=...`                   | Точка                                                                                    | `200`: очередь заданий с текущими версиями                                                                                                        |
| Сотрудник/водитель через Gateway → Fulfillment | `POST /v1/tasks/{taskId}/transitions`         | `action, expectedVersion, readyAt?, driverId?, deliveryEta?, reason?`                    | `200`: новое состояние и версия; действия `ACCEPT/REJECT/START/READY/ASSIGN_DRIVER/DISPATCH/COMPLETE/FAIL`; недопустимый переход - `409`          |
| UI через Gateway → Fulfillment                 | `POST /v1/routes`                             | `origin, destination, mode`                                                              | `200`: `provider, directionsUrl, durationSeconds?, trafficAvailable`; при недоступности провайдеров - адрес назначения и `trafficAvailable=false` |
| Владелец через Gateway → Stores                | `PATCH /v1/stores/{storeId}`                  | `expectedVersion, hours?, deliveryZone?, deliveryFee?, paymentMethods?`                  | `200`: новая версия; только своя точка                                                                                                            |
| Владелец через Gateway → Menu                  | `PATCH /v1/stores/{storeId}/menu/{productId}` | `expectedVersion, price?, available?`                                                    | `200`: новая версия; только своя точка                                                                                                            |
| Менеджер/владелец через Gateway → Menu         | `POST /v1/promotions`                         | `scope=NATIONAL/LOCAL, country, storeId?, startsAt, endsAt, productIds, discountPercent` | `201`: `promotionId`; национальная акция - право менеджера, локальная - владельца точки                                                           |

Menu определяет период акции по часовому поясу точки, полученному Orders из Stores по доверенному каналу. `expiresAt` расчета не выходит за конец примененной акции. Orders гарантирует цену полного расчета до `expiresAt`; после создания заказа сохраняет ее снимок. Изменение доступности после расчета может привести к отказу точки, но не к молчаливой замене состава или суммы.

`START` возможен только после `PreparationAuthorized`, `COMPLETE` - после `HandoverAuthorized`; доставка дополнительно требует назначенного водителя и состояния `DISPATCHED`. `FAIL` публикует отказ исполнения и передает ошибку в Orders для отмены и возврата денег. Изменение задания и его событие сохраняются одной транзакцией.

## 3. События

У всех событий одинаковые служебные поля. Пример сообщения:

```json
{
  "eventId": "7210d072-d9c3-43b4-80a7-c8bcf4c72a45",
  "eventType": "FulfillmentAccepted",
  "schemaVersion": 1,
  "occurredAt": "2026-09-07T12:00:00Z",
  "producer": "fulfillment",
  "aggregateId": "73193e64-74ed-45b1-ae86-c7e13779164b",
  "aggregateVersion": 2,
  "correlationId": "690b9a7d-f4fa-482a-8b57-312f9ec147bc",
  "payload": {
    "orderId": "73193e64-74ed-45b1-ae86-c7e13779164b",
    "storeId": "294695bd-52cc-45fc-90df-d2aee3c8e5be",
    "taskId": "2064fc27-a8e0-4da6-a9ac-34c553c25123",
    "readyAt": "2026-09-07T12:30:00Z",
    "reservationExpiresAt": "2026-09-07T12:10:00Z"
  }
}
```

| Событие | Издатель → подписчик | Обязательные поля payload | Реакция |
|---|---|---|---|
| `OrderPlaced` | Orders → Fulfillment | `orderId, storeId, items, fulfillmentType, paymentMethod, deliveryAddress?, customerContact` | Создать одно задание на заказ |
| `FulfillmentAccepted` | Fulfillment → Orders | `orderId, storeId, taskId, readyAt, reservationExpiresAt` | Сохранить время; начать оплату либо разрешить приготовление |
| `FulfillmentRejected` | Fulfillment → Orders | `orderId, reason` | Отклонить заказ без оплаты |
| `PaymentSucceeded` | Payments → Orders | `orderId, paymentId, total, method, paidAt` | Сверить сумму и заказ; разрешить приготовление и выдачу, либо вернуть позднюю оплату отмененного заказа |
| `PaymentFailed` | Payments → Orders | `orderId, paymentId, attemptId, reason` | Показать ошибку; новая попытка допустима до истечения резерва |
| `PreparationAuthorized` | Orders → Fulfillment | `orderId` | Разрешить приготовление принятого заказа |
| `HandoverAuthorized` | Orders → Fulfillment | `orderId, paymentId` | Разрешить выдачу после готовности |
| `FulfillmentUpdated` | Fulfillment → Orders | `orderId, status, readyAt, deliveryEta?, driverId?` | Обновить статус и ожидаемое время; состояния `PREPARING/READY/DISPATCHED/COMPLETED` |
| `FulfillmentFailed` | Fulfillment → Orders | `orderId, reason` | Отменить исполнение, при наличии оплаты запросить возврат |
| `OrderCancelled` | Orders → Fulfillment | `orderId, reason` | Прекратить исполнение и освободить резерв, не возобновлять по запоздавшему разрешению |
| `PaymentRefunded` | Payments → Orders | `orderId, paymentId, refundId, total` | Установить статус возврата `REFUNDED` |

Одно событие может прийти несколько раз (at least once). Сервис записывает его в таблицу outbox в той же транзакции, что и изменения заказа или платежа, затем отправляет в брокер. Получатель сохраняет `eventId` в inbox вместе с результатом обработки. Если такой ID уже есть, повторно выполнять действие не нужно. У каждого получателя своя очередь. После нескольких неудачных попыток сообщение попадает в отдельную очередь ошибок (DLQ), где его можно разобрать.

`aggregateVersion` задает порядок изменений одного объекта в одном сервисе. Версии из разных сервисов между собой не сравниваются. Если пропущено сообщение, обработка ждет его восстановления. Старое сообщение не должно менять статус обратно. Orders и Fulfillment проверяют, можно ли выполнить действие в текущем статусе. Отмененный заказ не запускается заново. Даже если отмена пришла раньше `OrderPlaced`, Fulfillment запоминает ее и не создает активное задание позже.

Orders хранит срок оплаты в БД и проверяет его в фоне, в том числе после перезапуска. Если оплата пришла после отмены, запускается возврат. Если платеж создан, но ответ потерялся, запрос повторяется с тем же ключом. Когда результат неизвестен, Payments сначала запрашивает его у платежного провайдера.

## 4. Внешние интеграции

Payments обращается к API платежного провайдера. При создании платежа получает его ID и ссылку на оплату. Результат приходит через webhook: ID платежа, статус и сумма. Payments проверяет подпись, сумму и валюту, а повторные уведомления игнорирует. Переход покупателя обратно на сайт еще не означает успешную оплату. Данные карты вводятся на странице провайдера и в BLT не хранятся.

Fulfillment адаптирует API как минимум двух картографических провайдеров: координаты начала и назначения → маршрут, длительность, признак актуальных данных о пробках. Используются таймаут и переключение провайдера. Нельзя выдавать оценку без пробок за точное время доставки; интерфейс явно отмечает отсутствие этих данных.

Контактные данные и адрес доступны только участникам исполнения соответствующего заказа. Их не следует добавлять в общие журналы; брокер разрешает чтение `OrderPlaced` только Fulfillment.
