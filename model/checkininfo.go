package model

// CheckinInfoWebConnectivity contains the array of URLs returned by the checkin API
type CheckinInfoWebConnectivity struct {
	Urls []URLInfo `json:"urls"`
}

// CheckinInfo contains the return test objects from the checkin API
type CheckinInfo struct {
	WebConnectivity CheckinInfoWebConnectivity `json:"web_connectivity"`
}
