package solana

// ClientOption is a function that allows to specify options for the client
type ClientOption func(*client) error

// WithMaxRetries sets the maximum number of retries for the engine when sending a transaction on-chain
func WithMaxRetries(maxRetries uint) ClientOption {
	return func(e *client) error {
		e.maxRetries = maxRetries
		return nil
	}
}

// WithClientRPC sets the Solana client RPC for the engine
func WithClientRPC(clientRPC rpcService) ClientOption {
	return func(e *client) error {
		e.clientRPC = clientRPC
		return nil
	}
}

// WithClientWS sets the Solana client WS for the engine
func WithClientWS(clientWS wsService) ClientOption {
	return func(e *client) error {
		e.clientWS = clientWS
		return nil
	}
}
