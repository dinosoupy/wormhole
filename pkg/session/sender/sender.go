package sender

import (
	"io"
	"sync"

	"github.com/pion/webrtc/v2"
	"github.com/dinosoupy/wormhole/pkg/stats"
	"github.com/dinosoupy/wormhole/pkg/session"
	"github.com/dinosoupy/wormhole/pkg/session/common"
)

const (
	// Must be <= 16384
	senderBuffSize = 16384
)

type outputMsg struct {
	n    int
	buff []byte
}

type SenderSession struct {
	session 	session.Session	
	stream      io.Reader
	initialized bool

	dataChannel *webrtc.DataChannel
	dataBuff    []byte
	msgToBeSent []outputMsg
	stopSending chan struct{}
	output      chan outputMsg

	doneCheckLock sync.Mutex
	doneCheck     bool

	// Stats/infos
	readingStats *stats.Stats
}

// Sender Session constructor
func Sender(c Config) *SenderSession {
	return &SenderSession{
		session:      session.New(nil, nil),
		stream:       c.Stream,
		initialized:  false,
		dataBuff:     make([]byte, senderBuffSize),
		stopSending:  make(chan struct{}, 1),
		output:       make(chan outputMsg, senderBuffSize*10),
		doneCheck:    false,
		readingStats: stats.New(),
	}
}  

// Config contains custom configuration for a session
type Config struct {
	common.Configuration
	Stream io.Reader // The Stream to read from
}