package session

import (
	"fmt"
	"io"
	"os"

	"github.com/pion/webrtc/v2"
	"github.com/dinosoupy/wormhole/pkg/stats"
	"github.com/dinosoupy/wormhole/pkg/session/common"
)

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

//New creates a new sender session
func NewSender(s session.Session, f io.Reader) *Session {
	return &Session{
		sess:         s,
		stream:       f,
		initialized:  false,
		dataBuff:     make([]byte, senderBuffSize),
		stopSending:  make(chan struct{}, 1),
		output:       make(chan outputMsg, senderBuffSize*10),
		doneCheck:    false,
		readingStats: stats.New(),
	}
}  

// New creates a new receiver session
func NewReceiver(f io.Reader) *Session {
	return new(session.New(nil, nil, ""), f)
}