package main

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	clients   = make(map[*websocket.Conn]*Client)
	clientsMu sync.Mutex
)

type Client struct {
	pc         *webrtc.PeerConnection
	videoTrack *webrtc.TrackLocalStaticRTP
	audioTrack *webrtc.TrackLocalStaticRTP
	videoSeq   uint16
	audioSeq   uint16
}

func main() {
	// Start RTP listeners for video and audio
	go func() {
		if err := startRTPListener(5000, "video"); err != nil {
			log.Fatalf("Failed to start video RTP listener: %v", err)
		}
	}()
	go func() {
		if err := startRTPListener(5002, "audio"); err != nil {
			log.Fatalf("Failed to start audio RTP listener: %v", err)
		}
	}()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("WebSocket upgrade failed: %v", err)
			return
		}
		defer func() {
			if err := conn.Close(); err != nil {
				log.Printf("WebSocket close error: %v", err)
			}
		}()

		// Create media engine with our codecs
		mediaEngine := &webrtc.MediaEngine{}
		if err := mediaEngine.RegisterCodec(
			webrtc.RTPCodecParameters{
				RTPCodecCapability: webrtc.RTPCodecCapability{
					MimeType:  "video/H264",
					ClockRate: 90000,
				},
				PayloadType: 96,
			},
			webrtc.RTPCodecTypeVideo,
		); err != nil {
			log.Printf("Failed to register video codec: %v", err)
			return
		}

		if err := mediaEngine.RegisterCodec(
			webrtc.RTPCodecParameters{
				RTPCodecCapability: webrtc.RTPCodecCapability{
					MimeType:  "audio/opus",
					ClockRate: 48000,
					Channels:  2,
				},
				PayloadType: 111,
			},
			webrtc.RTPCodecTypeAudio,
		); err != nil {
			log.Printf("Failed to register audio codec: %v", err)
			return
		}

		// Create API with media engine
		api := webrtc.NewAPI(webrtc.WithMediaEngine(mediaEngine))
		config := webrtc.Configuration{
			ICEServers: []webrtc.ICEServer{
				{
					URLs: []string{"stun:stun.l.google.com:19302"},
				},
			},
		}

		peerConnection, err := api.NewPeerConnection(config)
		if err != nil {
			log.Printf("Failed to create peer connection: %v", err)
			return
		}

		// Create video track
		videoTrack, err := webrtc.NewTrackLocalStaticRTP(
			webrtc.RTPCodecCapability{MimeType: "video/H264", ClockRate: 90000},
			"video",
			"stream",
		)
		if err != nil {
			log.Printf("Failed to create video track: %v", err)
			peerConnection.Close()
			return
		}

		// Create audio track
		audioTrack, err := webrtc.NewTrackLocalStaticRTP(
			webrtc.RTPCodecCapability{MimeType: "audio/opus", ClockRate: 48000, Channels: 2},
			"audio",
			"stream",
		)
		if err != nil {
			log.Printf("Failed to create audio track: %v", err)
			peerConnection.Close()
			return
		}

		// Add tracks to peer connection
		videoSender, err := peerConnection.AddTrack(videoTrack)
		if err != nil {
			log.Printf("Failed to add video track: %v", err)
			peerConnection.Close()
			return
		}

		audioSender, err := peerConnection.AddTrack(audioTrack)
		if err != nil {
			log.Printf("Failed to add audio track: %v", err)
			peerConnection.Close()
			return
		}

		// Read incoming RTCP packets from senders to prevent blocking
		go func() {
			rtcpBuf := make([]byte, 1500)
			for {
				if _, _, rtcpErr := videoSender.Read(rtcpBuf); rtcpErr != nil {
					return
				}
			}
		}()
		go func() {
			rtcpBuf := make([]byte, 1500)
			for {
				if _, _, rtcpErr := audioSender.Read(rtcpBuf); rtcpErr != nil {
					return
				}
			}
		}()

		// Generate random sequence numbers
		var videoSeq, audioSeq uint16
		if err := binary.Read(rand.Reader, binary.LittleEndian, &videoSeq); err != nil {
			log.Printf("Failed to generate video sequence number: %v", err)
			peerConnection.Close()
			return
		}
		if err := binary.Read(rand.Reader, binary.LittleEndian, &audioSeq); err != nil {
			log.Printf("Failed to generate audio sequence number: %v", err)
			peerConnection.Close()
			return
		}

		client := &Client{
			pc:         peerConnection,
			videoTrack: videoTrack,
			audioTrack: audioTrack,
			videoSeq:   videoSeq,
			audioSeq:   audioSeq,
		}

		// Store client
		clientsMu.Lock()
		clients[conn] = client
		clientsMu.Unlock()

		// Handle cleanup
		defer func() {
			clientsMu.Lock()
			delete(clients, conn)
			clientsMu.Unlock()
			if err := peerConnection.Close(); err != nil {
				log.Printf("Failed to close peer connection: %v", err)
			}
		}()

		// Create offer
		offer, err := peerConnection.CreateOffer(nil)
		if err != nil {
			log.Printf("Failed to create offer: %v", err)
			return
		}

		// Set local description
		if err = peerConnection.SetLocalDescription(offer); err != nil {
			log.Printf("Failed to set local description: %v", err)
			return
		}

		// Send offer
		if err = conn.WriteJSON(offer); err != nil {
			log.Printf("Failed to send offer: %v", err)
			return
		}

		// Wait for answer
		var answer webrtc.SessionDescription
		if err = conn.ReadJSON(&answer); err != nil {
			log.Printf("Failed to read answer: %v", err)
			return
		}

		// Set remote description
		if err = peerConnection.SetRemoteDescription(answer); err != nil {
			log.Printf("Failed to set remote description: %v", err)
			return
		}

		// Keep connection alive
		select {}
	})

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe("127.0.0.1:8080", nil); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

func startRTPListener(port int, mediaType string) error {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IP{127, 0, 0, 1}, Port: port})
	if err != nil {
		return fmt.Errorf("failed to listen on UDP port %d: %w", port, err)
	}
	defer conn.Close()

	log.Printf("Listening for %s RTP on port %d", mediaType, port)

	buffer := make([]byte, 1500)
	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("Failed to read from UDP: %v", err)
			continue
		}

		var pkt rtp.Packet
		if err := pkt.Unmarshal(buffer[:n]); err != nil {
			log.Printf("Failed to unmarshal RTP packet: %v", err)
			continue
		}

		clientsMu.Lock()
		for ws, client := range clients {
			var track *webrtc.TrackLocalStaticRTP
			var seq *uint16

			switch mediaType {
			case "video":
				track = client.videoTrack
				seq = &client.videoSeq
			case "audio":
				track = client.audioTrack
				seq = &client.audioSeq
			default:
				continue
			}

			newPkt := &rtp.Packet{
				Header: rtp.Header{
					Version:        2,
					PayloadType:    pkt.PayloadType,
					SequenceNumber: *seq,
					Timestamp:      pkt.Timestamp,
					SSRC:           pkt.SSRC,
				},
				Payload: pkt.Payload,
			}

			*seq++
			if err := track.WriteRTP(newPkt); err != nil {
				log.Printf("Failed to write RTP packet: %v", err)
				ws.Close()
				delete(clients, ws)
				continue
			}
		}
		clientsMu.Unlock()
	}
}
