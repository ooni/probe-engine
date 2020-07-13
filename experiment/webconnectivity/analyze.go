package webconnectivity

// AnalysisResult contains the results of the analysis performed on the
// client. We obtain it by comparing the measurement and the control.
type AnalysisResult struct {
	DNSConsistency  string  `json:"dns_consistency"`
	BodyLengthMatch *string `json:"body_length_match"`
	HeadersMatch    *string `json:"header_match"`
	StatusCodeMatch *bool   `json:"status_code_match"`
	TitleMatch      *bool   `json:"title_match"`
	Accessible      *bool   `json:"accessible"`
	Blocking        *string `json:"blocking"`
}

// Analyze performs follow-up analysis on the webconnectivity measurement by
// comparing the measurement (tk.TestKeys) and the control (tk.Control). This
// function will return the results of the analysis.
func Analyze(tk *TestKeys) (out AnalysisResult) {
	return
}
