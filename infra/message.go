package infra

// Message represents data in Cloud Functions from Cloud PubSub
type Message struct {
	Data []byte `json:"data"`
}
