package encoding

import (
	"github.com/VortexSilence/X/common/chacha20"
	"github.com/VortexSilence/X/common/codec"
	"github.com/VortexSilence/X/common/crypto"
)

type EZ struct {
	mes []byte
	k   []byte
}

func NewEncodeZ(mes []byte, k []byte) *EZ {
	return &EZ{
		mes: mes,
		k:   k,
	}
}

func (x *EZ) EnAES() *EZ {
	m, err := crypto.EncryptAES256CBC(x.mes, x.k)
	if err != nil {
		return nil
	}
	x.mes = m
	return x
}

func (x *EZ) EnCha() *EZ {
	m, err := chacha20.Encrypt(x.mes, x.k)
	if err != nil {
		return nil
	}
	x.mes = m
	return x
}

func (x *EZ) Compress() *EZ {
	x.mes = codec.CompressSnappy(x.mes)
	return x
}

func (x *EZ) Build() []byte {
	return x.mes
}

type DZ struct {
	mes []byte
	k   []byte
}

func NewDZ(mes []byte, k []byte) *DZ {
	return &DZ{
		mes: mes,
		k:   k,
	}
}

func (x *DZ) DeAES() *DZ {
	m, err := crypto.DecryptAES256CBC(x.mes, x.k)
	if err != nil {
		return nil
	}
	x.mes = m
	return x
}

func (x *DZ) DecodeCha() *DZ {
	m, err := chacha20.Decrypt(x.mes, x.k)
	if err != nil {
		return nil
	}
	x.mes = m
	return x
}

func (x *DZ) Decompress() *DZ {
	d, err := codec.DecompressSnappy(x.mes)
	if err != nil {
		return nil
	}
	x.mes = d
	return x
}

func (x *DZ) Build() []byte {
	return x.mes
}
