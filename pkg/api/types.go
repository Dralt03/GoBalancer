package api

type Backend struct {
	Address   string `json:"address"`
	Weight    int64  `json:"weight"`
	Alive     bool   `json:"alive"`
	ConnCount int64  `json:"conn_count"`
}

type AddBackendRequest struct {
	Address string `json:"address"`
	Weight  int64  `json:"weight"`
}

type UpdateWeightRequest struct {
	Address string `json:"address"`
	Weight  int64  `json:"weight"`
}