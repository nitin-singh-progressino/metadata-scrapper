package schema

type Metadata struct {
	Image       string `json:"image"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Secrets struct {
	NitinAccessKey string `json:"nitinaccesskey"`
	NitinSecret    string `json:"nitinsecret"`
}
