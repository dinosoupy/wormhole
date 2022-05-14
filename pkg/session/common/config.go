package common

import (
	"io"

	"github.com/dinosoupy/wormhole/pkg/session"
)

// Configuration common to both Sender and Receiver session
type Configuration struct {
	SDPProvider  io.Reader                 // The SDP reader
	SDPOutput    io.Writer                 // The SDP writer
	OnCompletion session.CompletionHandler // Handler to call on session completion
}
