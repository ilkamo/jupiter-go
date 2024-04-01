package solana

// ClientOption is a function that allows to specify options for the client.
type ClientOption func(*client) error

// WithMaxRetries sets the maximum number of retries for the engine when sending a transaction on-chain
func WithMaxRetries(maxRetries uint) ClientOption {
	return func(e *client) error {
		e.maxRetries = maxRetries
		return nil
	}
}

// WithClientRPC sets the RPC service for the client.
func WithClientRPC(clientRPC rpcService) ClientOption {
	return func(e *client) error {
		e.clientRPC = clientRPC
		return nil
	}
}

// MonitorOption is a function that allows to specify options for the monitor.
type MonitorOption func(*monitor) error

// WithMonitorSubscriber sets the subscriber service for the monitor.
func WithMonitorSubscriber(subscriber subscriberService) MonitorOption {
	return func(m *monitor) error {
		m.sub = subscriber
		return nil
	}
}
