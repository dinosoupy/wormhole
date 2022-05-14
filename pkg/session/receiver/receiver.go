package receiver

import (
	"io"

	"github.com/pion/webrtc/v2"
	"github.com/dinosoupy/wormhole/pkg/session"
	"github.com/dinosoupy/wormhole/pkg/session/common"
)

// Session is a receiver session
type ReceiverSession struct {
	session     session.Session
	stream      io.Writer
	msgChannel  chan webrtc.DataChannelMessage
	initialized bool
}

func Receiver(c Config) *ReceiverSession {
	return &ReceiverSession{
		session:     session.New(nil, nil),
		stream:      c.Stream,
		msgChannel:  make(chan webrtc.DataChannelMessage, 4096*2),
		initialized: false,
	}
}

// Config contains custom configuration for a session
type Config struct {
	common.Configuration
	Stream io.Writer // The Stream to write to
}
