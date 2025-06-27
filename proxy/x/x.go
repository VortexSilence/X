package x

import (
	"core/common/chacha20"
	"core/common/codec"
)

type EX struct {
	mes []byte
	k   []byte
}

func NewEncodeX(mes []byte, k []byte) *EX {
	return &EX{
		mes: mes,
		k:   k,
	}
}
func (x *EX) EnCha() *EX {
	m, err := chacha20.Encrypt(x.mes, x.k)
	if err != nil {
		return nil
	}
	x.mes = m
	return x
}

func (x *EX) Compress() *EX {
	x.mes = codec.CompressSnappy(x.mes)
	return x
}

func (x *EX) Build() []byte {
	return x.mes
}

type DX struct {
	mes []byte
	k   []byte
}

func NewDX(mes []byte, k []byte) *DX {
	return &DX{
		mes: mes,
		k:   k,
	}
}
func (x *DX) DesktopCha() *DX {
	m, err := chacha20.Decrypt(x.mes, x.k)
	if err != nil {
		return nil
	}
	x.mes = m
	return x
}

func (x *DX) Decompress() *DX {
	d, err := codec.DecompressSnappy(x.mes)
	if err != nil {
		return nil
	}
	x.mes = d
	return x
}

func (x *DX) Build() []byte {
	return x.mes
}
