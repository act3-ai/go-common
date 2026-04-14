package address

// Address represents a mailing address.
type Address struct {
	Street  string `json:"street"`         // Street address.
	City    string `json:"city"`           // City name.
	State   string `json:"state,omitzero"` // State/region name.
	Country string `json:"country"`        // Country name.
	ZIP     string `json:"zip,omitzero"`   // ZIP code.
}
