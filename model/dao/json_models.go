package dao

// Config - app config
type Config struct {
	ServiceAddress   string `json:"service_address"`
	Driver           string `json:"driver"`
	ConnectionString string `json:"connection_string"`
	PageLimit        int    `json:"page_limit"`
}

type ApiError struct {
	Error string `json:"error,omitempty"`
}

type JwtClaims struct {
	Id int64
}
