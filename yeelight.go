package main

import (
	"fmt"
	"net"
)

type YeelightClient struct {
	Address string
}

func (yc *YeelightClient) convertAndSendRgbMessage(val float32) {
	yc.sendRgbMessage(int(val * 16777215))
}

func (yc *YeelightClient) sendRgbMessage(color int) {
	id := "1"
	method := "set_rgb"
	params := fmt.Sprintf("%v, \"smooth\", 500", color)
	msg := "{\"id\":" + id + ",\"method\":\""
	msg += method + "\",\"params\":[" + params + "]}\r\n"

	tcpAddr, err := net.ResolveTCPAddr("tcp", yc.Address)
	if err != nil {
		println("ResolveTCPAddr failed:", err.Error())
		return
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		println("Dial failed:", err.Error())
		return
	}

	defer conn.Close()

	_, err = conn.Write([]byte(msg))
	if err != nil {
		println("Write to server failed:", err.Error())
		return
	}

	reply := make([]byte, 1024)

	_, err = conn.Read(reply)
	if err != nil {
		println("Write to server failed:", err.Error())
		return
	}

	println("reply from server=", string(reply))
}
