package shiftclaiming

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"cloud.google.com/go/firestore"
	cloudtaskss "github.com/yesaswi/shift-claiming-automation/internal/cloudtasks"
)

type Service struct {
    firestoreClient *firestore.Client
    cloudTasksClient *cloudtasks.Client
}

func NewService(firestoreClient *firestore.Client, cloudTasksClient *cloudtasks.Client) *Service {
    return &Service{
        firestoreClient: firestoreClient,
        cloudTasksClient: cloudTasksClient,
    }
}

func (s *Service) StartClaiming() error {
    fmt.Println(`{"message": "Starting shift claiming...", "severity": "info"}`)
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
    fmt.Println(`{"message": "Stopping shift claiming...", "severity": "info"}`)
    // Update the start/stop flag in Firestore
    configDoc := s.firestoreClient.Collection("configuration").Doc("config")
    _, err := configDoc.Set(context.Background(), map[string]interface{}{
        "startStopFlag": false,
    })
    if err != nil {
        return fmt.Errorf("failed to update start/stop flag: %v", err)
    }

    // Delete all pending tasks from the queue and stop claiming
    err = cloudtaskss.DeleteAllTasks(s.cloudTasksClient, "autoclaimer-42", "us-east4", "barbequeue")
    if err != nil {
        return fmt.Errorf("failed to delete pending tasks: %v", err)
    }

    return nil
}

func (s *Service) ScheduleClaimTask(scheduleTime time.Time) error {
    // Schedule a new task to trigger the /claim endpoint
    fmt.Printf(`{"message": "Scheduling claim task...", "schedule_time": "%s", "severity": "info"}`+"\n", scheduleTime)
    _, err := cloudtaskss.CreateTask(s.cloudTasksClient, "autoclaimer-42", "us-east4", "barbequeue", "https://autoclaimer-h5km45tdpq-uk.a.run.app/claim", scheduleTime)
    if err != nil {
        return fmt.Errorf("failed to schedule claim task: %v", err)
    }
    return nil
}

func (s *Service) ClaimShift() error {
    fmt.Println(`{"message": "Claiming shift...", "severity": "info"}`)

    // Check if claiming is enabled
    configDoc := s.firestoreClient.Collection("configuration").Doc("config")
    configDocSnap, err := configDoc.Get(context.Background())
    if err != nil {
        return fmt.Errorf("failed to retrieve configuration: %v", err)
    }
    configData := configDocSnap.Data()
    startStopFlag, ok := configData["startStopFlag"].(bool)
    if !ok {
        return fmt.Errorf("missing or invalid 'startStopFlag' in configuration")
    }
    if !startStopFlag {
        fmt.Println(`{"message": "Claiming is disabled", "severity": "warning"}`)
        return nil
    }

    // Retrieve the claiming configuration from Firestore
    authConfigDoc := s.firestoreClient.Collection("configuration").Doc("auth")
    authConfigDocSnap, err := authConfigDoc.Get(context.Background())
    if err != nil {
        return fmt.Errorf("failed to retrieve claiming configuration: %v", err)
    }
    var authConfig map[string]interface{}
    err = authConfigDocSnap.DataTo(&authConfig)
    if err != nil {
        return fmt.Errorf("failed to parse auth configuration: %v", err)
    }

    // Retrieve the claiming configuration from Firestore
    shiftConfigDoc := s.firestoreClient.Collection("configuration").Doc("shiftconfig")
    shiftConfigDocSnap, err := shiftConfigDoc.Get(context.Background())
    if err != nil {
        return fmt.Errorf("failed to retrieve claiming configuration: %v", err)
    }
    var shiftConfig map[string]interface{}
    err = shiftConfigDocSnap.DataTo(&shiftConfig)
    if err != nil {
        return fmt.Errorf("failed to parse claiming configuration: %v", err)
    }

    // Extract the necessary configuration values
    cookie, ok := authConfig["cookie"].(string)
    if !ok {
        return fmt.Errorf("missing or invalid 'cookie' in claiming configuration")
    }
    xAPIToken, ok := authConfig["x_api_token"].(string)
    if !ok {
        return fmt.Errorf("missing or invalid 'x_api_token' in claiming configuration")
    }

    shiftStartDate, ok := shiftConfig["shift_start_date"].(string)
    if !ok {
        return fmt.Errorf("missing or invalid 'shift_start_date' in claiming configuration")
    }
    shiftRange, ok := shiftConfig["shift_range"].(string)
    if !ok {
        return fmt.Errorf("missing or invalid 'shift_range' in claiming configuration")
    }
    // A, B, C1, C2
    shiftGroup, ok := shiftConfig["shift_group"].(string)
    if !ok {
        return fmt.Errorf("missing or invalid 'shift_group' in claiming configuration")
    }

    userID, ok := authConfig["user_id"].(string)
    if !ok {
        return fmt.Errorf("missing or invalid 'user_id' in claiming configuration")
    }

    // Fetch available shifts
    availableShifts, err := fetchAvailableShifts(cookie, xAPIToken, shiftStartDate, shiftRange)
    if err != nil {
        if strings.HasPrefix(err.Error(), "Swap list disabled") ||
            strings.HasPrefix(err.Error(), "Please wait") ||
            strings.HasPrefix(err.Error(), "Session Timeout") {
            fmt.Printf(`{"message": "Claiming is disabled", "error": "%s", "severity": "info"}`+"\n", err.Error())
            
            if strings.HasPrefix(err.Error(), "Swap list disabled") {
                // Schedule the next claim task after 30 minutes
                err = s.ScheduleClaimTask(time.Now().Add(30 * time.Minute))
                if err != nil {
                    return fmt.Errorf("failed to schedule next claim task: %v", err)
                }
            } else if strings.HasPrefix(err.Error(), "Please wait") {
                // Schedule the next claim task after 3 seconds
                err = s.ScheduleClaimTask(time.Now().Add(3 * time.Second))
                if err != nil {
                    return fmt.Errorf("failed to schedule next claim task: %v", err)
                }
            } else if strings.HasPrefix(err.Error(), "Session Timeout") {
                // Lof the error and stop claiming
                fmt.Printf(`{"message": "Session Timeout. Please sign in again.", "severity": "alert"}`+"\n")
            }
            
            return nil
        }
        return fmt.Errorf("failed to fetch available shifts: %v", err)
    }

    // Schedule the next claim task
    err = s.ScheduleClaimTask(time.Now().Add(5 * time.Second))
    if err != nil {
        return fmt.Errorf("failed to schedule next claim task: %v", err)
    }

    if len(availableShifts) == 0 {
        fmt.Println(`{"message": "No available shifts to claim", "severity": "info"}`)
        // To a new random documentID in the requests collection
        s.firestoreClient.Collection("requests").NewDoc().Set(context.Background(), map[string]interface{}{
            "timestamp": time.Now(),
            "message": "No available shifts to claim",
        })
        return nil
    }

    for _, shift := range availableShifts {
    s.firestoreClient.Collection("available_shifts").NewDoc().Set(context.Background(), map[string]interface{}{
            "timestamp": time.Now(),
            "schId": shift.SchId,
            "locId": shift.LocId,
            "stnName": shift.StnName,
            "date": shift.Date,
            "hours": shift.Hours,
            "shiftGroup": shift.ShiftGroup,
            "start": shift.Start,
            "end": shift.End,
        })
    }

    // Claim the shifts
    claimingResults := claimShifts(availableShifts, cookie, xAPIToken, userID, shiftStartDate, shiftGroup)
    for _, result := range claimingResults {
        s.firestoreClient.Collection("claims").NewDoc().Set(context.Background(), map[string]interface{}{
            "timestamp": result.Timestamp,
            "shiftId": result.ShiftID,
            "claimingStatus": result.ClaimingStatus,
        })
    }
    if len(claimingResults) == 0 {
        fmt.Println(`{"message": "No shifts claimed", "severity": "alert"}`)
    } else if len(claimingResults) < len(availableShifts) {
        fmt.Printf(`{"message": "Some shifts failed to claim", "claiming_results": %v, "severity": "alert"}`+"\n", claimingResults)
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
            fmt.Printf(`{"message": "Failed to close response body", "error": "%v", "severity": "warning"}`+"\n", err)
        }
    }(resp.Body)
    if resp.StatusCode != http.StatusOK {
        bodyBytes, err := io.ReadAll(resp.Body)
        if err != nil {
            return nil, fmt.Errorf("failed to read response body: %v", err)
        }
        bodyString := string(bodyBytes)
        if strings.Contains(bodyString, "Swap list disabled.  (30) minutes idle required for reset.") {
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

func claimShifts(shifts []Shift, cookie, xAPIToken string, userID string, shiftStartDate string, shiftGroup string) []ClaimingResult {
    var claimingResults []ClaimingResult
    claimingURL := "https://tmwork.net/api/shift/swap/claim"
    req, err := http.NewRequest("PUT", claimingURL, nil)
    if err != nil {
        fmt.Printf(`{"message": "Failed to create claiming request", "error": "%v", "severity": "error"}`+"\n", err)
    }
    client := &http.Client{}
    q := req.URL.Query()
	q.Add("id", userID)
    q.Add("bid", "3557")
    req.Header.Set("Cookie", cookie)
    req.Header.Set("X-API-Token", xAPIToken)
    for _, shift := range shifts {
        shiftDate, _ := time.Parse("2006-01-02", shift.Date)
        compareDate, _ := time.Parse("2006-01-02", shiftStartDate)
        if  shiftDate.Before(compareDate) && !strings.Contains(shiftGroup, shift.ShiftGroup) {
            continue
        }
        q.Add("schid", fmt.Sprintf("%d", shift.SchId))
        req.URL.RawQuery = q.Encode()
        resp, err := client.Do(req)
        if err != nil {
            fmt.Printf(`{"message": "Failed to claim shift", "shift_id": %d, "error": "%v", "severity": "error"}`+"\n", shift.SchId, err)
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
                fmt.Printf(`{"message": "Failed to close response body", "error": "%v", "severity": "warning"}`+"\n", err)
            }
        }(resp.Body)
        if resp.StatusCode != http.StatusOK {
            bodyBytes, err := io.ReadAll(resp.Body)
            if err != nil {
                fmt.Printf(`{"message": "Failed to read response body", "error": "%v", "severity": "error"}`+"\n", err)
            } else {
                bodyString := string(bodyBytes)
                fmt.Printf(`{"message": "Failed to claim shift", "shift_id": %d, "status_code": %d, "response": "%s", "severity": "error"}`+"\n", shift.SchId, resp.StatusCode, bodyString)
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
