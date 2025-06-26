package codec

func Encode(msg []byte) []byte {
	return buildFinalMySQLLikePacket(msg)
}
func Decode(msg []byte) ([]byte, error) {
	return decodeFinalMySQLLikePacket(msg)
}
