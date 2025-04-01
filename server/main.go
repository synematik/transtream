package main

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

var (
	// Global slice of video tracks for all connected clients.
	peersMu sync.Mutex
	peers   []*webrtc.TrackLocalStaticRTP

	// Latest keyframe RTP packets (for simplicity, we cache one packet).
	keyframeMu     sync.Mutex
	latestKeyframe []*rtp.Packet
)

// Signal represents the JSON payload for SDP exchange.
type Signal struct {
	SDP  string `json:"sdp"`
	Type string `json:"type"`
}

// offerHandler handles incoming SDP offers, creates a PeerConnection with a video track,
// and immediately sends any cached keyframe so a late joiner can start decoding.
func offerHandler(w http.ResponseWriter, r *http.Request) {
	var offer Signal
	if err := json.NewDecoder(r.Body).Decode(&offer); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create a new PeerConnection.
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a video track using H264 codec.
	track, err := webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264},
		"video", "pion",
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Add the track to the PeerConnection.
	if _, err = peerConnection.AddTrack(track); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Save track in global list for broadcasting.
	addTrack(track)

	// Immediately send the latest keyframe (if available) to help new clients start decoding.
	keyframeMu.Lock()
	if latestKeyframe != nil {
		for _, pkt := range latestKeyframe {
			// WriteRTP may block; in production consider doing this asynchronously.
			if err := track.WriteRTP(pkt); err != nil {
				log.Println("Error sending cached keyframe:", err)
			}
		}
	}
	keyframeMu.Unlock()

	// Set the remote SDP offer.
	sdpOffer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  offer.SDP,
	}
	if err = peerConnection.SetRemoteDescription(sdpOffer); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create and set the local SDP answer.
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err = peerConnection.SetLocalDescription(answer); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send the answer back as JSON.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(peerConnection.LocalDescription())
}

// addTrack appends a new track to the list of peers.
func addTrack(track *webrtc.TrackLocalStaticRTP) {
	peersMu.Lock()
	defer peersMu.Unlock()
	peers = append(peers, track)
}

// isKeyframe performs a basic check to see if an RTP packet contains a keyframe.
// For H264, we check if the NAL unit type is 5 (IDR) or a fragmented FU-A with start bit and type 5.
func isKeyframe(packet *rtp.Packet) bool {
	if len(packet.Payload) < 1 {
		return false
	}
	naluType := packet.Payload[0] & 0x1F
	if naluType == 5 {
		return true
	}
	// Check for fragmented NAL (FU-A)
	if naluType == 28 && len(packet.Payload) > 1 {
		fuHeader := packet.Payload[1]
		if fuHeader&0x80 != 0 { // start of fragmented NAL unit
			origNALUType := packet.Payload[0] & 0x1F
			if origNALUType == 5 {
				return true
			}
		}
	}
	return false
}

// broadcast reads RTP packets from FFmpeg (from stdin) and sends them to every connected track.
// It also caches keyframe packets for new clients.
func broadcast() {
	reader := bufio.NewReader(os.Stdin)
	buf := make([]byte, 1500) // typical MTU size

	for {
		n, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				log.Println("FFmpeg stream ended.")
				return
			}
			log.Println("Error reading from FFmpeg:", err)
			continue
		}

		packet := &rtp.Packet{}
		if err := packet.Unmarshal(buf[:n]); err != nil {
			log.Println("Error unmarshaling RTP packet:", err)
			continue
		}

		// If the packet is a keyframe, cache it.
		if isKeyframe(packet) {
			keyframeMu.Lock()
			// In a real scenario, you might need to cache all RTP packets that compose the keyframe.
			latestKeyframe = []*rtp.Packet{packet}
			keyframeMu.Unlock()
		}

		// Broadcast the packet to all connected peers.
		peersMu.Lock()
		for _, t := range peers {
			if err := t.WriteRTP(packet); err != nil {
				log.Println("Error writing RTP packet to track:", err)
			}
		}
		peersMu.Unlock()
	}
}

func main() {
	http.HandleFunc("/offer", offerHandler)
	go broadcast()

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
