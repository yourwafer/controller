package client

import (
	"fmt"
	"testing"
	"time"
	"xa.com/manager/server/socket/core"
)

func TestCheckSum(t *testing.T) {
	negVal := []int8{-1, -2, -3, -4, -5, 0, 1, 2, 3, 4}
	buf := make([]byte, 0, len(negVal))
	for _, v := range negVal {
		buf = append(buf, byte(v))
	}
	fmt.Println(buf)
	check := CheckSum(buf)
	fmt.Println(check)
}

func TestCheckSum2(t *testing.T) {
	var mod int16 = -1
	header := &core.Header{Format: 1, State: 0, Sn: 0, Session: 0, Module: uint16(mod), Command: 18}
	bytes := header.Encode()
	check := CheckSum(bytes)
	for _, i := range bytes {
		fmt.Println(int64(int8(i)))
	}
	fmt.Println(check)
}

func TestClient_Write(t *testing.T) {
	address := "192.168.11.192:11111"
	param := "18015685055086616"
	cli, err := Connect(address, 5*time.Second)
	if err != nil {
		fmt.Println("连接失败")
		return
	}
	var module int16 = -30
	req := struct {
		Module  int16  `json:"module"`
		Command int16  `json:"command"`
		Param   string `json:"param"`
	}{Module: 0, Command: 2, Param: param}
	cli.Write(uint16(module), 5, req)
	time.Sleep(1 * time.Second)
	cli.Write(uint16(module), 5, req)
	time.Sleep(2 * time.Second)
}

func TestChannelIsClosed(t *testing.T) {
	ch := make(chan int, 1)
	ch <- 2
	chClosed := false
	fmt.Println("step 1")
	select {
	case <-ch:
		chClosed = false
	default:
		chClosed = true
	}
	fmt.Println("step 2")
	older := <-ch
	fmt.Println("step 3")
	fmt.Println(older)
	fmt.Println(chClosed)
}
