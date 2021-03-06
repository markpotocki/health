package models

type ClientInfo struct {
	CName string `json:"name"`
	CPort int    `json:"port"`
	CURL  string `json:"url"`
	Key   string `json:"key"`
}

func (ci ClientInfo) Name() string {
	return ci.CName
}

func (ci ClientInfo) URL() string {
	return ci.CURL
}
