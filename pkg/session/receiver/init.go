package receiver

import (
	"fmt"

	"github.com/pion/webrtc/v2"
	log "github.com/sirupsen/logrus"
)

// Initialize creates the connection, the datachannel and creates the  offer
func (s *ReceiverSession) Initialize() error {
	if s.initialized {
		return nil
	}
	if err := s.session.CreateConnection(s.onConnectionStateChange()); err != nil {
		log.Errorln(err)
		return err
	}
	s.createDataHandler()
	if err := s.session.ReadSDP(); err != nil {
		log.Errorln(err)
		return err
	}
	if err := s.session.CreateAnswer(); err != nil {
		log.Errorln(err)
		return err
	}

	s.initialized = true
	return nil
}

// Start initializes the connection and the file transfer
func (s *ReceiverSession) Start() error {
	if err := s.Initialize(); err != nil {
		return err
	}

	// Handle data
	s.receiveData()
	s.session.OnCompletion()
	return nil
}

func (s *ReceiverSession) createDataHandler() {
	s.session.OnDataChannel(func(d *webrtc.DataChannel) {
		log.Debugf("New DataChannel %s %d\n", d.Label(), d.ID())
		s.session.NetworkStats.Start()
		d.OnMessage(s.onMessage())
		d.OnClose(s.onClose())
	})
}

func (s *ReceiverSession) receiveData() {
	log.Infoln("Starting to receive data...")
	defer log.Infoln("Stopped receiving data...")

	// Consume the message channel, until done
	// Does not stop on error
	for {
		select {
		case <-s.session.Done:
			s.session.NetworkStats.Stop()
			fmt.Printf("\nNetwork: %s\n", s.session.NetworkStats.String())
			return
		case msg := <-s.msgChannel:
			n, err := s.stream.Write(msg.Data)

			if err != nil {
				log.Errorln(err)
			} else {
				currentSpeed := s.session.NetworkStats.Bandwidth()
				fmt.Printf("Transferring at %.2f MB/s\r", currentSpeed)
				s.session.NetworkStats.AddBytes(uint64(n))
			}
		}
	}
}
