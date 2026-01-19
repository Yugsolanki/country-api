package models

// Response returned by our API
type Country struct {
	Name       string `json:"name"`
	Capital    string `json:"capital"`
	Currency   string `json:"currency"`
	Population int64  `json:"population"`
}

// Response the response from REST Countries API
type RestCountryResponse struct {
	Name       NameInfo            `json:"name"`
	Capital    []string            `json:"capital"`
	Currencies map[string]Currency `json:"currencies"`
	Population int64               `json:"population"`
}

type NameInfo struct {
	Common   string `json:"common"`
	Official string `json:"official"`
}

type Currency struct {
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
}
