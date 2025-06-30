package pipe

import (
	"github.com/VortexSilence/X/config"
	"github.com/VortexSilence/X/transport/http"
)

func HandlePipe(mes []byte) []byte {
	c := config.Get()
	if c.Pipe == "http" {
		return (&http.HTTPCamouflage{}).Wrap(mes)
	}
	return mes
}
