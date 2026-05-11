package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/CR45-NITT/cr45-reduced/backend/internal/auth"
	"github.com/CR45-NITT/cr45-reduced/backend/internal/domain"
	"github.com/CR45-NITT/cr45-reduced/backend/internal/store"
)

type Server struct {
	store store.Repository
	auth  *auth.Service
}

func NewServer(s store.Repository, authService *auth.Service) *Server {
	return &Server{store: s, auth: authService}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/api/auth/login", s.handleLogin)
	mux.HandleFunc("/api/timetable", s.handleGetTimetable)
	mux.HandleFunc("/api/admin/override", s.handleUpdateOverride)
	mux.HandleFunc("/api/admin/slot", s.handleDeleteSlot)
	return mux
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *Server) handleGetTimetable(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	classID := strings.TrimSpace(r.URL.Query().Get("class_id"))
	if classID == "" {
		writeJSONError(w, http.StatusBadRequest, "class_id is required")
		return
	}

	timetable, err := s.store.GetResolvedTimetable(r.Context(), classID)
	if err != nil {
		if errors.Is(err, store.ErrClassNotFound) {
			writeJSONError(w, http.StatusNotFound, store.ErrClassNotFound.Error())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "failed to resolve timetable")
		return
	}
	writeJSON(w, http.StatusOK, timetable)
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	contentType := strings.TrimSpace(r.Header.Get("Content-Type"))
	if !strings.HasPrefix(contentType, "application/json") {
		writeJSONError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
		return
	}

	var req loginRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	token, err := s.auth.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, store.ErrInvalidLogin) {
			writeJSONError(w, http.StatusUnauthorized, store.ErrInvalidLogin.Error())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "login failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (s *Server) handleUpdateOverride(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	token, err := auth.BearerToken(r.Header.Get("Authorization"))
	if err != nil || s.auth.Validate(token) != nil {
		writeJSONError(w, http.StatusUnauthorized, store.ErrUnauthorized.Error())
		return
	}

	contentType := strings.TrimSpace(r.Header.Get("Content-Type"))
	if !strings.HasPrefix(contentType, "application/json") {
		writeJSONError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
		return
	}

	var req domain.UpdateOverrideRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.ClassID = strings.TrimSpace(req.ClassID)
	req.CourseCode = strings.TrimSpace(req.CourseCode)
	req.StartTime = strings.TrimSpace(req.StartTime)
	req.EndTime = strings.TrimSpace(req.EndTime)
	req.Venue = strings.TrimSpace(req.Venue)
	req.Status = strings.TrimSpace(req.Status)

	if req.ClassID == "" || req.CourseCode == "" || req.StartTime == "" || req.EndTime == "" || req.Venue == "" || req.Status == "" {
		writeJSONError(w, http.StatusBadRequest, "all fields are required")
		return
	}

	if err := s.store.UpsertOverride(r.Context(), req); err != nil {
		switch {
		case errors.Is(err, store.ErrClassNotFound):
			writeJSONError(w, http.StatusNotFound, store.ErrClassNotFound.Error())
		case errors.Is(err, store.ErrInvalidSlotIdx):
			writeJSONError(w, http.StatusBadRequest, store.ErrInvalidSlotIdx.Error())
		default:
			writeJSONError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleDeleteSlot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	token, err := auth.BearerToken(r.Header.Get("Authorization"))
	if err != nil || s.auth.Validate(token) != nil {
		writeJSONError(w, http.StatusUnauthorized, store.ErrUnauthorized.Error())
		return
	}

	classID := strings.TrimSpace(r.URL.Query().Get("class_id"))
	slotIndexText := strings.TrimSpace(r.URL.Query().Get("slot_index"))
	if classID == "" || slotIndexText == "" {
		writeJSONError(w, http.StatusBadRequest, "class_id and slot_index are required")
		return
	}

	var req domain.DeleteSlotRequest
	req.ClassID = classID
	if _, err := fmt.Sscanf(slotIndexText, "%d", &req.SlotIndex); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid slot_index")
		return
	}

	if err := s.store.DeleteSlot(r.Context(), req); err != nil {
		switch {
		case errors.Is(err, store.ErrClassNotFound):
			writeJSONError(w, http.StatusNotFound, store.ErrClassNotFound.Error())
		case errors.Is(err, store.ErrInvalidSlotIdx):
			writeJSONError(w, http.StatusBadRequest, store.ErrInvalidSlotIdx.Error())
		default:
			writeJSONError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
