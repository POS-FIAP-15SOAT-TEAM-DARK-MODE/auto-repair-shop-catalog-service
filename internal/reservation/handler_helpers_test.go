package reservation

import "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/messaging"

func messagingEnvelope(serviceOrderID string) messaging.Envelope {
	return messaging.Envelope{MessageID: "b7a1c6f0-2f64-4f6e-9a49-0d5a0b1e6c11", Type: "ReleaseStock", ServiceOrderID: serviceOrderID}
}
