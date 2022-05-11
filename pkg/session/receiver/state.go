package receiver

import (
	"github.com/pion/webrtc/v2"
	log "github.com/sirupsen/logrus"
)

func (s *ReceiverSession) onConnectionStateChange() func(connectionState webrtc.ICEConnectionState) {
	return func(connectionState webrtc.ICEConnectionState) {
		log.Infof("ICE Connection State has changed: %s\n", connectionState.String())
	}
}

func (s *ReceiverSession) onMessage() func(msg webrtc.DataChannelMessage) {
	return func(msg webrtc.DataChannelMessage) {
		// Store each message in the message channel
		s.msgChannel <- msg
	}
}

func (s *ReceiverSession) onClose() func() {
	return func() {
		close(s.session.Done)
	}
}
