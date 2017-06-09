package moetcpserver

import "fmt"

type Msg interface {
	fmt.Stringer // String()
	Encode() (b []byte)
	Decode(bs []byte) Msg
}
