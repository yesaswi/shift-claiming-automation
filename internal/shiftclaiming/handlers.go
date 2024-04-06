package shiftclaiming

import (
	"encoding/json"
	"net/http"

	customerrors "github.com/yesaswi/shift-claiming-automation/pkg/errors"
)

func (s *Service) HandleStartCommand(w http.ResponseWriter, r *http.Request) {
	err := s.StartClaiming()
	if err != nil {
		httpErr := customerrors.LogAndReturnError(s.log, err, "Failed to start claiming", "ERROR", http.StatusInternalServerError)
		http.Error(w, httpErr.Error(), httpErr.(customerrors.HTTPError).StatusCode)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Shift claiming started successfully"))
}

func (s *Service) HandleStopCommand(w http.ResponseWriter, r *http.Request) {
	err := s.StopClaiming()
	if err != nil {
		httpErr := customerrors.LogAndReturnError(s.log, err, "Failed to stop claiming", "ERROR", http.StatusInternalServerError)
		http.Error(w, httpErr.Error(), httpErr.(customerrors.HTTPError).StatusCode)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Shift claiming stopped successfully"))
}

func (s *Service) HandleClaimCommand(w http.ResponseWriter, r *http.Request) {
	err := s.ClaimShift()
	if err != nil {
		httpErr := customerrors.LogAndReturnError(s.log, err, "Failed to claim shift", "ERROR", http.StatusInternalServerError)
		http.Error(w, httpErr.Error(), httpErr.(customerrors.HTTPError).StatusCode)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Triggered shift claiming"))
}

func (s *Service) HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *Service) HandleAuthentication(w http.ResponseWriter, r *http.Request) {
    var loginData LoginData

    err := json.NewDecoder(r.Body).Decode(&loginData)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Pass the individual fields to the Authenticate method
    err = s.Authenticate(loginData.LoginURL, loginData.LoginCode, loginData.Username, loginData.Password)
    if err != nil {
        http.Error(w, err.Error(), http.StatusUnauthorized)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
}

type LoginData struct {
	LoginURL  string `json:"login_url"`
	LoginCode string `json:"login_code"`
	Username  string `json:"username"`
	Password  string `json:"password"`
}
