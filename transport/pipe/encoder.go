package pipe

import "github.com/VortexSilence/X/proxy/x"

func HandlePipeEncoder(mes []byte) []byte {
	return x.NewEncodeX(mes, []byte("0123456789abcdef0123456789abcdef")).
		EnCha().
		Compress().
		Build()
}

func HandlePipeDecoder(mes []byte) []byte {
	return x.NewDX(mes, []byte("0123456789abcdef0123456789abcdef")).
		Decompress().
		DesktopCha().
		Build()
}
