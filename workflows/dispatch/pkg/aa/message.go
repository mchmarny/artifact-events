package aa

// VulnerabilityMessage is the payload of a vulnerability message.
type VulnerabilityMessage struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
}
