package pipe

import (
	"github.com/VortexSilence/X/config"
	"github.com/VortexSilence/X/transport/http"
)

type ICamouflage interface {
	Wrap(buf []byte, proto string) []byte
	Unwrap(data []byte) (string, []byte, error)
	WrapResponse(data []byte, protocol string, statusCode int) []byte
	UnwrapResponse(data []byte) (string, []byte, error)
}

func HandlePipe() ICamouflage {
	//TODO add config and impl all function in https or other tr
	_ = config.Get()
	// if c.Pipe == "http" {
	return (&http.HTTPCamouflage{})
	// } else if c.Pipe == "https" {
	// 	// return (&https.TLSCamouflage{})
	// }
	// return mes
}
