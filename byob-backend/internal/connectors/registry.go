package connectors

import (
	"fmt"

	"github.com/GauravS11112003/BYOB/byob-backend/internal/config"
)

// Build constructs a Connector from a single connector configuration entry,
// dispatching on the connector type.
func Build(cfg config.ConnectorConfig) (Connector, error) {
	switch cfg.Type {
	case "rest":
		return NewRESTConnector(RESTConfig{
			Name:     cfg.Name,
			URL:      cfg.URL,
			Interval: cfg.PollInterval.AsDuration(),
		}), nil
	case "kafka":
		return NewKafkaConnector(KafkaConfig{
			Name:    cfg.Name,
			Brokers: cfg.Brokers,
			Topic:   cfg.Topic,
			GroupID: cfg.GroupID,
		}), nil
	default:
		return nil, fmt.Errorf("unknown connector type %q (connector %q)", cfg.Type, cfg.Name)
	}
}

// BuildAll constructs every connector declared in the configuration list.
func BuildAll(cfgs []config.ConnectorConfig) ([]Connector, error) {
	out := make([]Connector, 0, len(cfgs))
	for _, c := range cfgs {
		conn, err := Build(c)
		if err != nil {
			return nil, err
		}
		out = append(out, conn)
	}
	return out, nil
}
