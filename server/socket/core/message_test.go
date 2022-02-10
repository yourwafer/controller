package core

import (
	"testing"
)

func TestHeader_Encode(t *testing.T) {
	var module int16 = -10
	header := Header{Format: 1, State: 123456789, Sn: 987654321, Session: 123467899, Module: uint16(module), Command: uint16(module)}
	buf := header.Encode()
	if len(buf) != 25 {
		t.Fatalf("消息头长度错误25!=%d\n", len(buf))
	}
	headerDec, _ := HeaderDecode(buf)
	if header != *headerDec {
		t.Fatalf("需要%v，但得到%v\n", header, headerDec)
	}
}

func TestMessage_Encode(t *testing.T) {
	header := Header{Format: 1, State: 123456789, Sn: 987654321, Session: 123467899, Module: 1, Command: 2}
	msg := Message{Header: &header, Body: []byte{0, 1, 2, 3}, Attachment: make([]byte, 0)}
	buf := msg.Encode()
	shouldGetLen := 25 + 4 + 4
	if len(buf) != shouldGetLen {
		t.Errorf("序列化失败，长度错误%d!=%d\n", shouldGetLen, len(buf))
	}
	msgDec := MessageDecode(buf)
	if *msg.Header != *msgDec.Header {
		t.Errorf("反序列化失败")
	}
	if len(msg.Body) != len(msgDec.Body) {
		t.Fatalf("反序列化失败")
		return
	}
	for i := len(msg.Body) - 1; i >= 0; i-- {
		if msgDec.Body[i] != msg.Body[i] {
			t.Fatalf("反序列化失败")
			return
		}
	}
}
