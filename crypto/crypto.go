package crypto

type PublicKey interface {
	ToBytes() []byte
	FromBytes(bz []byte) error
}
