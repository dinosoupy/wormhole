package sender

import (
	"fmt"

	"github.com/pion/webrtc/v2"
	log "github.com/sirupsen/logrus"
)

func (s *SenderSession) onConnectionStateChange() func(connectionState webrtc.ICEConnectionState) {
	return func(connectionState webrtc.ICEConnectionState) {
		log.Infof("ICE Connection State has changed: %s\n", connectionState.String())
		if connectionState == webrtc.ICEConnectionStateDisconnected {
			s.stopSending <- struct{}{}
		}
	}
}

func (s *SenderSession) onOpenHandler() func() {
	return func() {
		s.session.NetworkStats.Start()

		log.Infof("Starting to send data...")
		defer log.Infof("Stopped sending data...")

		s.writeToNetwork()
	}
}

func (s *SenderSession) onCloseHandler() func() {
	return func() {
		s.close(true)
	}
}

func (s *SenderSession) close(calledFromCloseHandler bool) {
	if !calledFromCloseHandler {
		s.dataChannel.Close()
	}

	// Sometime, onCloseHandler is not invoked, so it's a work-around
	s.doneCheckLock.Lock()
	if s.doneCheck {
		s.doneCheckLock.Unlock()
		return
	}
	s.doneCheck = true
	s.doneCheckLock.Unlock()
	s.dumpStats()
	close(s.session.Done)
}

func (s *SenderSession) dumpStats() {
	fmt.Printf(`
Disk   : %s
Network: %s
`, s.readingStats.String(), s.session.NetworkStats.String())
}