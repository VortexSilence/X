package pipe

import (
	"core/config"
	"core/transport/http"
)

func HandlePipe(mes []byte) []byte {
	c := config.Get()
	if c.Pipe == "http" {
		return (&http.HTTPCamouflage{
			Host:      "pashmak.com",
			Path:      "/api/v1/data",
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		}).Wrap(mes)
	}
	return mes
}
