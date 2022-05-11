package main

import (
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/examples/internal/signal"
)

func main() {
	// HTTP server for saving SDP strings (base64 encoded)
	sdpChan := signal.HTTPSDPServer()

	// Decoding SD from base64 and storing it in SessionDescription struct
	offer := webrtc.SessionDescription{}
	signal.Decode(<-sdpChan, &offer)
	fmt.Println("")

	// Config for ICE servers stored as ICEServer json struct
	peerConnectionConfig := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}
	
}