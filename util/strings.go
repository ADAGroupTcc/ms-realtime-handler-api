package util

const (
	UserIsConnected                          = "websocket_handler: user_id: %s connected in this pod_name: %s\n"
	ErrorTypeErr                             = "error"
	NumberOfActiveConnections                = "websocket_handler: number of active connections: %d\n"
	FailedToUpgradeConnection                = "websocket_handler: failed to upgrade connection"
	UnableToParseEventResponse               = "unable to parser eventToReceiver response"
	UnableToParseWsEventResponse             = "unable to parser websocket event response"
	ReceiverNotOnlineInPod                   = "receiver_id %s is not online in this pod_name: %s\n"
	ConnectionClosedUnexpectedly             = "websocket_handler: connection closed unexpectedly"
	ConnectionClosed                         = "websocket_handler: connection closed"
	FailedToCreateHttpClient                 = "Failed to create http client"
	FailedToStartServer                      = "Failed to start server"
	FailedToLoadEnvVars                      = "Failed to load environment variables"
	FailedToReadMessageFromWebsocket         = "websocket_handler: failed to read message from webSocket client"
	FailedToUnmarshalMessage                 = "websocket_handler: failed to unmarshal message"
	FailedToValidateMessage                  = "websocket_handler: failed to validate message"
	FailedToPublishMessageToPubSubBroker     = "websocket_handler: failed to publish message to pubsub broker"
	PublishMessageToPubSubBrokerSuccessfully = "websocket_handler: message type %s published successfully\n"
	ErrorToInitInstrumentation               = "error to initialize instrumentation"
)
