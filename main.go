package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

type wakeRequest struct {
	MAC              string `json:"mac"`
	BroadcastAddress string `json:"broadcast_address,omitempty"`
	Port             int    `json:"port,omitempty"`
}

type wakeResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

var macRegex = regexp.MustCompile(`^([0-9A-Fa-f]{2}[:-]){5}[0-9A-Fa-f]{2}$`)

func normalizeMAC(mac string) ([]byte, error) {
	if !macRegex.MatchString(mac) {
		return nil, fmt.Errorf("invalid MAC address format: %s", mac)
	}
	clean := strings.ReplaceAll(strings.ReplaceAll(mac, ":", ""), "-", "")
	return hex.DecodeString(clean)
}

func buildMagicPacket(macBytes []byte) []byte {
	packet := make([]byte, 0, 102)
	// 6 bytes of 0xFF
	for i := 0; i < 6; i++ {
		packet = append(packet, 0xFF)
	}
	// MAC repeated 16 times
	for i := 0; i < 16; i++ {
		packet = append(packet, macBytes...)
	}
	return packet
}

func sendMagicPacket(macBytes []byte, broadcastAddr string, port int) error {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", broadcastAddr, port))
	if err != nil {
		return fmt.Errorf("resolve udp addr: %w", err)
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return fmt.Errorf("dial udp: %w", err)
	}
	defer conn.Close()

	packet := buildMagicPacket(macBytes)
	_, err = conn.Write(packet)
	if err != nil {
		return fmt.Errorf("write packet: %w", err)
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func wakeHandler(defaultBroadcast string, defaultPort int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, wakeResponse{Status: "error", Message: "method not allowed"})
			return
		}

		var req wakeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, wakeResponse{Status: "error", Message: "invalid json body"})
			return
		}

		macBytes, err := normalizeMAC(req.MAC)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, wakeResponse{Status: "error", Message: err.Error()})
			return
		}

		broadcast := req.BroadcastAddress
		if broadcast == "" {
			broadcast = defaultBroadcast
		}
		port := req.Port
		if port == 0 {
			port = defaultPort
		}

		if err := sendMagicPacket(macBytes, broadcast, port); err != nil {
			log.Printf("failed to send magic packet to %s: %v", req.MAC, err)
			writeJSON(w, http.StatusInternalServerError, wakeResponse{Status: "error", Message: err.Error()})
			return
		}

		log.Printf("magic packet sent: mac=%s broadcast=%s port=%d", req.MAC, broadcast, port)
		writeJSON(w, http.StatusOK, wakeResponse{Status: "ok", Message: fmt.Sprintf("magic packet sent to %s via %s:%d", req.MAC, broadcast, port)})
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, wakeResponse{Status: "ok"})
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	listenAddr := envOr("LISTEN_ADDR", ":8089")
	defaultBroadcast := envOr("DEFAULT_BROADCAST", "255.255.255.255")
	defaultPort := 9
	if v := os.Getenv("DEFAULT_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &defaultPort)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/wake", wakeHandler(defaultBroadcast, defaultPort))
	mux.HandleFunc("/healthz", healthHandler)

	srv := &http.Server{
		Addr:              listenAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	log.Printf("wol-relay listening on %s (default broadcast=%s port=%d)", listenAddr, defaultBroadcast, defaultPort)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}