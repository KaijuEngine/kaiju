package master_server

type Error = uint8

const (
	ErrorNone = Error(iota)
	ErrorIncorrectPassword
	ErrorServerDoesntExist
)
