package main

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

var upgrader = websocket.Upgrader{}

func main() {
	// Create a MediaEngine with default codecs (H264 for video and Opus for audio)
	m := &webrtc.MediaEngine{}
	if err := m.RegisterDefaultCodecs(); err != nil {
		log.Fatal(err)
	}

	api := webrtc.NewAPI(webrtc.WithMediaEngine(m))

	// HTTP server: serve static files from "./static" (for our HTML page)
	http.Handle("/", http.FileServer(http.Dir("./")))

	// WebSocket endpoint for signaling
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Upgrade error:", err)
			return
		}
		defer conn.Close()

		// Create a new PeerConnection
		peerConnection, err := api.NewPeerConnection(webrtc.Configuration{})
		if err != nil {
			log.Println("Error creating PeerConnection:", err)
			return
		}
		defer peerConnection.Close()

		// Create local video track (H264) and audio track (Opus)
		videoTrack, err := webrtc.NewTrackLocalStaticRTP(
			webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264},
			"video", "pion",
		)
		if err != nil {
			log.Println("Error creating video track:", err)
			return
		}
		audioTrack, err := webrtc.NewTrackLocalStaticRTP(
			webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus},
			"audio", "pion",
		)
		if err != nil {
			log.Println("Error creating audio track:", err)
			return
		}
		if _, err = peerConnection.AddTrack(videoTrack); err != nil {
			log.Println("Error adding video track:", err)
			return
		}
		if _, err = peerConnection.AddTrack(audioTrack); err != nil {
			log.Println("Error adding audio track:", err)
			return
		}

		// ICE candidate exchange: send each candidate to the browser via WebSocket
		peerConnection.OnICECandidate(func(c *webrtc.ICECandidate) {
			if c == nil {
				return
			}
			candidate := c.ToJSON()
			candidateMsg, _ := json.Marshal(candidate)
			conn.WriteMessage(websocket.TextMessage, candidateMsg)
		})

		// Wait for the offer from the browser
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading offer:", err)
			return
		}
		var offer webrtc.SessionDescription
		if err = json.Unmarshal(msg, &offer); err != nil {
			log.Println("Error unmarshalling offer:", err)
			return
		}
		if err = peerConnection.SetRemoteDescription(offer); err != nil {
			log.Println("Error setting remote description:", err)
			return
		}

		// Create and send an answer back to the client
		answer, err := peerConnection.CreateAnswer(nil)
		if err != nil {
			log.Println("Error creating answer:", err)
			return
		}
		if err = peerConnection.SetLocalDescription(answer); err != nil {
			log.Println("Error setting local description:", err)
			return
		}
		answerMsg, _ := json.Marshal(answer)
		conn.WriteMessage(websocket.TextMessage, answerMsg)

		// Start a goroutine to read RTP packets from UDP for the video track (port 5004)
		go func() {
			videoConn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: 5004})
			if err != nil {
				log.Println("Error listening for video RTP:", err)
				return
			}
			defer videoConn.Close()
			buf := make([]byte, 1500)
			for {
				n, _, err := videoConn.ReadFrom(buf)
				if err != nil {
					if err == io.EOF {
						break
					}
					log.Println("Error reading video RTP:", err)
					continue
				}
				packet := &rtp.Packet{}
				if err := packet.Unmarshal(buf[:n]); err != nil {
					log.Println("Error unmarshalling video RTP packet:", err)
					continue
				}
				if err = videoTrack.WriteRTP(packet); err != nil {
					log.Println("Error writing video RTP to track:", err)
				}
			}
		}()

		// Start a goroutine to read RTP packets from UDP for the audio track (port 5006)
		go func() {
			audioConn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: 5006})
			if err != nil {
				log.Println("Error listening for audio RTP:", err)
				return
			}
			defer audioConn.Close()
			buf := make([]byte, 1500)
			for {
				n, _, err := audioConn.ReadFrom(buf)
				if err != nil {
					if err == io.EOF {
						break
					}
					log.Println("Error reading audio RTP:", err)
					continue
				}
				packet := &rtp.Packet{}
				if err := packet.Unmarshal(buf[:n]); err != nil {
					log.Println("Error unmarshalling audio RTP packet:", err)
					continue
				}
				if err = audioTrack.WriteRTP(packet); err != nil {
					log.Println("Error writing audio RTP to track:", err)
				}
			}
		}()

		// Keep the WebSocket connection open
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				log.Println("WebSocket closed:", err)
				break
			}
		}
	})

	log.Println("Server started at :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
