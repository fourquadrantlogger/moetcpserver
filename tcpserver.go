package moetcpserver

import (
	"fmt"
	"net"
)

func InitTCP(addr string) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", addr) //获取一个tcpAddr
	if err != nil {
		panic(err)
	}

	listener, err := net.ListenTCP("tcp", tcpAddr) //监听一个端口
	if err != nil {
		panic(err)
	}
	fmt.Println("Tcpserver start accept,at ", addr)
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println("-------建立了链接-------", conn.RemoteAddr().String())

		Conns.Set(conn.RemoteAddr().String(), Client_Con_io{Con: conn})
		cli, _ := Conns.Get(conn.RemoteAddr().String())
		go cli.Init()
	}

}
