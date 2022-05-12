package session

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/pion/webrtc/v2"
	"github.com/dinosoupy/wormhole/pkg/stats"
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
func Sender(f io.Reader) *SenderSession {
	return &SenderSession{
		session:      session.New(nil, nil),
		stream:       f,
		initialized:  false,
		dataBuff:     make([]byte, senderBuffSize),
		stopSending:  make(chan struct{}, 1),
		output:       make(chan outputMsg, senderBuffSize*10),
		doneCheck:    false,
		readingStats: stats.New(),
	}
}  