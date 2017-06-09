package moetcpserver

import (
	"fmt"

	"net"

	"git.oschina.net/quxueche2016/DriversTrainingTimerecordingSystem/Config"
)

type Client_Con_io struct {
	Con              net.Conn
	returnsignal     chan uint16
	waitingreturnmap map[uint16]Msg
	MsgChan_Up       chan Msg
	MsgChan_Down     chan Msg
}

func (this *Client_Con_io) Init() {
	this.returnsignal = make(chan uint16, 1)
	this.MsgChan_Up = make(chan Msg, 100)
	this.MsgChan_Down = make(chan Msg, 100)
	go this.up()
	go this.down()
}

func (this *Client_Con_io) up() {
	for {

		defer func() { // 必须要先声明defer，否则不能捕获到panic异常
			if err := recover(); err != nil {
				fmt.Println("panic ", err) // 这里的err其实就是panic传入的内容，55
				this.Con.Close()
				Conns.Delete(this.Con.RemoteAddr().String())
			}
		}()

		//region up
		buf := make([]byte, Config.BufferLen)

		l, err := this.Con.Read(buf)
		if err != nil {
			panic(err)
		}
		fmt.Println("receive from ", this.Con.RemoteAddr().String(), "len=", len(buf[:l]), buf[:l])

		msg, err := new(Msg).Decode(buf[:l])
		if err != nil {
			panic(err)
		}

		this.MsgChan_Up <- *msg
		//endregion

	}
}

func (this *Client_Con_io) down() {
	for {
		defer func() { // 必须要先声明defer，否则不能捕获到panic异常
			if err := recover(); err != nil {
				fmt.Println("panic ", err) // 这里的err其实就是panic传入的内容，55
				this.Con.Close()
				Conns.Delete(this.Con.RemoteAddr().String())
			}
		}()

		m := <-this.MsgChan_Down
		bs := m.Encode()
		n, err := this.Con.Write(bs)
		if err != nil {
			panic("connect error")
		}
		//
		if n != len(bs) {
			panic("connect error")
		}
		fmt.Println("send to ", this.Con.RemoteAddr().String(), "len=", len(bs), bs)
	}
}
