package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/shepard-labs/go-us-uk-english-translator/translator"
)

type ConvertRequest struct {
	Text   string `json:"text"`
	Target string `json:"target,omitempty"` // "american" or "british". Defaults to "american".
}

type ConvertResponse struct {
	OriginalText     string `json:"original_text"`
	ConvertedText    string `json:"converted_text"`
	ReplacementsMade int    `json:"replacements_made"`
	TargetDirection  string `json:"target_direction"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func main() {
	port := flag.Int("port", 8080, "Port to listen on")
	flag.Parse()

	http.HandleFunc("/v1/convert", handleConvert)

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Starting API server on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func handleConvert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Be lenient with Content-Type if they append charset
	contentType := r.Header.Get("Content-Type")
	if contentType != "" && !strings.HasPrefix(strings.ToLower(contentType), "application/json") {
		writeError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
		return
	}

	var req ConvertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	target := strings.ToLower(strings.TrimSpace(req.Target))
	if target == "" {
		target = "american"
	}

	var dir translator.Direction
	switch target {
	case "american":
		dir = translator.DirectionAmerican
	case "british":
		dir = translator.DirectionBritish
	default:
		writeError(w, http.StatusBadRequest, "Invalid target direction, must be 'american' or 'british'")
		return
	}

	converter, err := translator.NewConverter(dir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to initialize converter")
		return
	}

	convertedText, replacementsMade := converter.Convert(req.Text, "api")

	resp := ConvertResponse{
		OriginalText:     req.Text,
		ConvertedText:    convertedText,
		ReplacementsMade: replacementsMade,
		TargetDirection:  target,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Error: msg})
}
