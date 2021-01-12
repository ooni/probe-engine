package model

// CheckInInfoWebConnectivity contains the array of URLs returned by the checkin API
type CheckInInfoWebConnectivity struct {
	URLs []URLInfo `json:"urls"`
}

// CheckInInfo contains the return test objects from the checkin API
type CheckInInfo struct {
	ReportID        string                      `json:"report_id"`
	WebConnectivity *CheckInInfoWebConnectivity `json:"web_connectivity"`
}