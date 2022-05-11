package receiver

import (
	"io"

	"github.com/pion/webrtc/v2"
	"github.com/dinosoupy/wormhole/pkg/session/common"
)

// Session is a receiver session
type ReceiverSession struct {
	session     session.Session
	stream      io.Writer
	msgChannel  chan webrtc.DataChannelMessage
	initialized bool
}

func Receiver(f io.Writer) *ReceiverSession {
	return &ReceiverSession{
		session:     session.New(nil, nil),
		stream:      f,
		msgChannel:  make(chan webrtc.DataChannelMessage, 4096*2),
		initialized: false,
	}
}
