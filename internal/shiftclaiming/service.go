package shiftclaiming

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"cloud.google.com/go/firestore"
	cloudtaskss "github.com/yesaswi/shift-claiming-automation/internal/cloudtasks"
	"go.uber.org/zap"
)

type Service struct {
    log             *zap.Logger
    firestoreClient *firestore.Client
    cloudTasksClient *cloudtasks.Client
}

func NewService(log *zap.Logger, firestoreClient *firestore.Client, cloudTasksClient *cloudtasks.Client) *Service {
    return &Service{
        log:             log,
        firestoreClient: firestoreClient,
        cloudTasksClient: cloudTasksClient,
    }
}

func (s *Service) StartClaiming() error {
    s.log.Info("Starting shift claiming...")
    // Update the start/stop flag in Firestore
    configDoc := s.firestoreClient.Collection("configuration").Doc("config")
    _, err := configDoc.Set(context.Background(), map[string]interface{}{
        "startStopFlag": true,
    })
    if err != nil {
        return fmt.Errorf("failed to update start/stop flag: %v", err)
    }

    // Schedule the initial claim task
    err = s.ScheduleClaimTask(time.Now())
    if err != nil {
        return fmt.Errorf("failed to schedule initial claim task: %v", err)
    }

    return nil
}

func (s *Service) StopClaiming() error {
    s.log.Info("Stopping shift claiming...")
    // Update the start/stop flag in Firestore
    configDoc := s.firestoreClient.Collection("configuration").Doc("config")
    _, err := configDoc.Set(context.Background(), map[string]interface{}{
        "startStopFlag": false,
    })
    if err != nil {
        return fmt.Errorf("failed to update start/stop flag: %v", err)
    }
    return nil
}

func (s *Service) ScheduleClaimTask(scheduleTime time.Time) error {
    // Schedule a new task to trigger the /claim endpoint
    s.log.Info("Scheduling claim task...", zap.Time("schedule_time", scheduleTime))
    _, err := cloudtaskss.CreateTask(s.cloudTasksClient, "autoclaimer-42", "us-east4", "barbequeue", "https://autoclaimer-h5km45tdpq-uk.a.run.app/claim", scheduleTime)
    if err != nil {
        return fmt.Errorf("failed to schedule claim task: %v", err)
    }
    return nil
}

func (s *Service) ClaimShift() error {
    s.log.Info("Claiming shift...")

    // Retrieve the claiming configuration from Firestore
    configDoc := s.firestoreClient.Collection("configuration").Doc("shiftconfig")
    var config map[string]interface{}
    docSnap, err := configDoc.Get(context.Background())
    if err != nil {
        return fmt.Errorf("failed to retrieve claiming configuration: %v", err)
    }
    err = docSnap.DataTo(&config)
    if err != nil {
        return fmt.Errorf("failed to parse claiming configuration: %v", err)
    }

    // Extract the necessary configuration values
    cookie, ok := config["cookie"].(string)
    if !ok {
        return fmt.Errorf("missing or invalid 'cookie' in claiming configuration")
    }
    xAPIToken, ok := config["x_api_token"].(string)
    if !ok {
        return fmt.Errorf("missing or invalid 'x_api_token' in claiming configuration")
    }
    shiftStartDate, ok := config["shift_start_date"].(string)
    if !ok {
        return fmt.Errorf("missing or invalid 'shift_start_date' in claiming configuration")
    }
    shiftRange, ok := config["shift_range"].(string)
    if !ok {
        return fmt.Errorf("missing or invalid 'shift_range' in claiming configuration")
    }
    userID, ok := config["user_id"].(string)
    if !ok {
        return fmt.Errorf("missing or invalid 'user_id' in claiming configuration")
    }

    // Fetch available shifts
    availableShifts, err := fetchAvailableShifts(cookie, xAPIToken, shiftStartDate, shiftRange)
    if err != nil {
        if strings.HasPrefix(err.Error(), "Swap list disabled") ||
            strings.HasPrefix(err.Error(), "Please wait") ||
            strings.HasPrefix(err.Error(), "Session Timeout") {
            return err
        }
        return fmt.Errorf("failed to fetch available shifts: %v", err)
    }

    // Schedule the next claim task
    err = s.ScheduleClaimTask(time.Now().Add(5 * time.Second))
    if err != nil {
        return fmt.Errorf("failed to schedule next claim task: %v", err)
    }

    if len(availableShifts) == 0 {
        s.log.Info("No available shifts to claim")
        return nil
    }

    // Claim the shifts
    claimingResults := claimShifts(availableShifts, cookie, xAPIToken, userID)
    if len(claimingResults) == 0 {
        s.log.Info("No shifts claimed")
    } else if len(claimingResults) < len(availableShifts) {
        s.log.Info("Some shifts failed to claim", zap.Any("claiming_results", claimingResults))
    }

    return nil
}

func fetchAvailableShifts(cookie, xAPIToken, shiftStartDate, shiftRange string) ([]Shift, error) {
    shiftListingURL := "https://tmwork.net/api/shift/swapboard"
    req, err := http.NewRequest("GET", shiftListingURL, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %v", err)
    }
    req.Header.Set("Cookie", cookie)
    req.Header.Set("X-API-Token", xAPIToken)
    q := req.URL.Query()
    q.Add("date", shiftStartDate)
    q.Add("range", shiftRange)
    req.URL.RawQuery = q.Encode()
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch shift listings: %v", err)
    }
    defer func(Body io.ReadCloser) {
        err := Body.Close()
        if err != nil {
            log.Printf("Failed to close response body: %v", err)
        }
    }(resp.Body)
    if resp.StatusCode != http.StatusOK {
        bodyBytes, err := io.ReadAll(resp.Body)
        if err != nil {
            return nil, fmt.Errorf("failed to read response body: %v", err)
        }
        bodyString := string(bodyBytes)
        if strings.Contains(bodyString, "Swap list disabled. (30) minutes idle required for reset.") {
            return nil, fmt.Errorf("%s", bodyString)
        } else if strings.Contains(bodyString, "Please wait [3] seconds to refresh list.") {
            return nil, fmt.Errorf("%s", bodyString)
        } else if strings.Contains(bodyString, "Session Timeout. Please sign in again.") {
            return nil, fmt.Errorf("%s", bodyString)
        }
        return nil, fmt.Errorf("failed to fetch shift listings: %s", bodyString)
    }

    var responseData []map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
        return nil, fmt.Errorf("failed to parse shift listings: %v", err)
    }
    var availableShifts []Shift
    for _, shiftData := range responseData {
        shift := Shift{
            SchId:      int(shiftData["SchId"].(float64)),
            LocId:      int(shiftData["LocId"].(float64)),
            StnName:    shiftData["StnName"].(string),
            Date:       shiftData["Date"].(string),
            Hours:      shiftData["Hours"].(float64),
            ShiftGroup: shiftData["ShiftGroup"].(string),
            Start:      shiftData["Start"].(string),
            End:        shiftData["End"].(string),
        }
        availableShifts = append(availableShifts, shift)
    }
    return availableShifts, nil
}

func claimShifts(shifts []Shift, cookie, xAPIToken string, userID string) []ClaimingResult {
    var claimingResults []ClaimingResult
    claimingURL := "https://tmwork.net/api/shift/swap/claim"
    req, err := http.NewRequest("PUT", claimingURL, nil)
    if err != nil {
        log.Printf("Failed to create claiming request: %v", err)
    }
    client := &http.Client{}
    q := req.URL.Query()
	q.Add("id", userID)
    q.Add("bid", "3557")
    req.Header.Set("Cookie", cookie)
    req.Header.Set("X-API-Token", xAPIToken)
    for _, shift := range shifts {
        q.Add("schid", fmt.Sprintf("%d", shift.SchId))
        req.URL.RawQuery = q.Encode()
        resp, err := client.Do(req)
        if err != nil {
            log.Printf("Failed to claim shift: %d, error: %v", shift.SchId, err)
            claimingResults = append(claimingResults, ClaimingResult{
                ShiftID:        fmt.Sprintf("%d", shift.SchId),
                ClaimingStatus: "failed",
                Timestamp:      time.Now(),
            })
            continue
        }
        defer func(Body io.ReadCloser) {
            err := Body.Close()
            if err != nil {
                log.Printf("Failed to close response body: %v", err)
            }
        }(resp.Body)
        if resp.StatusCode != http.StatusOK {
            bodyBytes, err := io.ReadAll(resp.Body)
            if err != nil {
                log.Printf("Failed to read response body: %v", err)
            } else {
                bodyString := string(bodyBytes)
                log.Printf("Failed to claim shift: %d, status code: %d, response: %s", shift.SchId, resp.StatusCode, bodyString)
            }
            claimingResults = append(claimingResults, ClaimingResult{
                ShiftID:        fmt.Sprintf("%d", shift.SchId),
                ClaimingStatus: "failed",
                Timestamp:      time.Now(),
            })
            continue
        }
        claimingResults = append(claimingResults, ClaimingResult{
            ShiftID:        fmt.Sprintf("%d", shift.SchId),
            ClaimingStatus: "success",
            Timestamp:      time.Now(),
        })
    }
    return claimingResults
}

type Shift struct {
    SchId      int     `json:"SchId"`
    LocId      int     `json:"LocId"`
    StnName    string  `json:"StnName"`
    Date       string  `json:"Date"`
    Hours      float64 `json:"Hours"`
    ShiftGroup string  `json:"ShiftGroup"`
    Start      string  `json:"Start"`
    End        string  `json:"End"`
}

type ClaimingResult struct {
    ShiftID        string    `json:"shift_id"`
    ClaimingStatus string    `json:"claiming_status"`
    Timestamp      time.Time `json:"timestamp"`
}
