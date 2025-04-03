package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

var (
	// Protects the list of active tracks.
	tracksMu sync.RWMutex
	// Global list of WebRTC tracks where RTP packets will be forwarded.
	tracks []*webrtc.TrackLocalStaticRTP
)

// rtpReceiver listens on UDP port 5004 for incoming RTP packets from FFmpeg.
func rtpReceiver() {
	addr, err := net.ResolveUDPAddr("udp", ":5004")
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	buf := make([]byte, 1500)
	for {
		n, _, err := conn.ReadFrom(buf)
		if err != nil {
			log.Println("Error reading from UDP:", err)
			continue
		}
		packet := &rtp.Packet{}
		if err := packet.Unmarshal(buf[:n]); err != nil {
			log.Println("Error unmarshalling RTP packet:", err)
			continue
		}
		// Broadcast the RTP packet to every active WebRTC track.
		tracksMu.RLock()
		for _, track := range tracks {
			if err := track.WriteRTP(packet); err != nil {
				log.Println("Error writing RTP to track:", err)
			}
		}
		tracksMu.RUnlock()
	}
}

// Offer represents an SDP offer from the client.
type Offer struct {
	SDP string `json:"sdp"`
}

// Answer represents the SDP answer returned to the client.
type Answer struct {
	SDP string `json:"sdp"`
}

// offerHandler handles the signaling exchange: it receives an SDP offer,
// creates a PeerConnection with a new video track, sets the remote description,
// and returns an SDP answer.
func offerHandler(w http.ResponseWriter, r *http.Request) {
	var offer Offer
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(body, &offer); err != nil {
		http.Error(w, "failed to parse JSON", http.StatusBadRequest)
		return
	}

	// Create a new PeerConnection.
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		http.Error(w, "failed to create peer connection", http.StatusInternalServerError)
		return
	}

	// Create a video track. (MIME type must match what FFmpeg outputs, e.g. video/H264.)
	track, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{
		MimeType: "video/H264",
	}, "video", "pion")
	if err != nil {
		http.Error(w, "failed to create track", http.StatusInternalServerError)
		return
	}

	// Add the track to the PeerConnection.
	_, err = peerConnection.AddTrack(track)
	if err != nil {
		http.Error(w, "failed to add track", http.StatusInternalServerError)
		return
	}

	// Add this track to the global list so that RTP packets are forwarded to it.
	tracksMu.Lock()
	tracks = append(tracks, track)
	tracksMu.Unlock()

	// Set the remote SDP offer.
	offerSD := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  offer.SDP,
	}
	if err := peerConnection.SetRemoteDescription(offerSD); err != nil {
		log.Println(err)
		http.Error(w, "failed to set remote description", http.StatusInternalServerError)
		return
	}

	// Create the SDP answer.
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		http.Error(w, "failed to create answer", http.StatusInternalServerError)
		return
	}
	if err = peerConnection.SetLocalDescription(answer); err != nil {
		http.Error(w, "failed to set local description", http.StatusInternalServerError)
		return
	}

	// Wait for ICE gathering to complete.
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	<-gatherComplete

	// Return the SDP answer.
	resp := Answer{
		SDP: peerConnection.LocalDescription().SDP,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	// Cleanup: Remove the track when the connection closes.
	peerConnection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		log.Println("Connection state has changed:", state.String())
		if state == webrtc.PeerConnectionStateClosed || state == webrtc.PeerConnectionStateFailed {
			tracksMu.Lock()
			for i, t := range tracks {
				if t == track {
					tracks = append(tracks[:i], tracks[i+1:]...)
					break
				}
			}
			tracksMu.Unlock()
		}
	})
}

func main() {
	// Start the RTP receiver in a separate goroutine.
	go rtpReceiver()

	// Serve static files (the HTML page) from the ./static folder.
	http.Handle("/", http.FileServer(http.Dir("./")))
	// HTTP endpoint for WebRTC signaling.
	http.HandleFunc("/offer", offerHandler)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
