package contracts

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewEnvelopeUsesSagaAsCorrelationID(t *testing.T) {
	envelope, err := NewEnvelope(MessageChargePaymentRequested, "order-1", "message-0", ChargePayment{
		OrderID: "order-1", OperationID: "order:order-1:payment", UserID: 42, Amount: 1000,
	})

	require.NoError(t, err)
	require.NotEmpty(t, envelope.MessageID)
	require.Equal(t, "order-1", envelope.SagaID)
	require.Equal(t, "order-1", envelope.CorrelationID)
	require.Equal(t, "message-0", envelope.CausationID)
	require.False(t, envelope.OccurredAt.IsZero())

	var command ChargePayment
	require.NoError(t, json.Unmarshal(envelope.Payload, &command))
	require.Equal(t, int64(1000), command.Amount)
}
