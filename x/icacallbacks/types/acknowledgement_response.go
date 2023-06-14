package types

type AckResponseStatus int

const (
	AckResponseStatusSuccess AckResponseStatus = iota
	AckResponseStatusTimeout
	AckResponseStatusFailure
)

type AcknowledgementResponse struct {
	Status       AckResponseStatus
	MsgResponses [][]byte
	Error        string
}
