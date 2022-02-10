package core

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"strconv"
)

/**
消息头
*/
type Header struct {
	Format  byte   // 格式
	State   uint32 // 状态
	Sn      uint64 // 序号
	Session uint64 // 会话标识
	Module  uint16 // 指令-模块号
	Command uint16 // 指令-方法
}

type Message struct {
	*Header
	Body       []byte
	Attachment []byte
}

func (msg *Message) ToString() string {
	res := make([]byte, 30)
	toString := msg.Header.ToString()
	res = append(res, []byte(toString)...)
	res = append(res, ';')
	//res = append(res, []byte("body:[")...)
	//res = strconv.AppendQuote(res, string(msg.Body))
	//res = append(res, []byte("]")...)
	return string(res)
}

func (header *Header) IsRequest() bool {
	return (header.State & 1) == 0
}

func (header *Header) Encode() []byte {
	buf := make([]byte, 25)
	buf[0] = header.Format
	binary.BigEndian.PutUint32(buf[1:], header.State)
	binary.BigEndian.PutUint64(buf[5:], header.Sn)
	binary.BigEndian.PutUint64(buf[13:], header.Session)
	binary.BigEndian.PutUint16(buf[21:], header.Module)
	binary.BigEndian.PutUint16(buf[23:], header.Command)
	return buf
}

func (header *Header) ToString() string {
	res := make([]byte, 10)
	res = strconv.AppendQuote(res, "Format")
	res = append(res, ':')
	res = strconv.AppendInt(res, int64(header.Format), 10)
	res = append(res, ',')
	res = strconv.AppendQuote(res, "State")
	res = append(res, ':')
	res = strconv.AppendInt(res, int64(header.State), 10)
	res = append(res, ',')
	res = strconv.AppendQuote(res, "Sn")
	res = append(res, ':')
	res = strconv.AppendInt(res, int64(header.Sn), 10)
	res = append(res, ',')
	res = strconv.AppendQuote(res, "Session")
	res = append(res, ':')
	res = strconv.AppendInt(res, int64(header.Session), 10)
	res = append(res, ',')
	res = strconv.AppendQuote(res, "Module")
	res = append(res, ':')
	res = strconv.AppendInt(res, int64(int16(header.Module)), 10)
	res = append(res, ',')
	res = strconv.AppendQuote(res, "Command")
	res = append(res, ':')
	res = strconv.AppendInt(res, int64(int16(header.Command)), 10)
	return string(res)
}

func (header *Header) Error() uint32 {
	return header.State >> 20
}

func HeaderDecode(buf []byte) (*Header, int32) {
	format := buf[0]
	state := binary.BigEndian.Uint32(buf[1:])
	sn := binary.BigEndian.Uint64(buf[5:])
	session := binary.BigEndian.Uint64(buf[13:])
	module := binary.BigEndian.Uint16(buf[21:])
	command := binary.BigEndian.Uint16(buf[23:])
	return &Header{Format: format, State: state, Sn: sn, Session: session, Module: module, Command: command}, 25
}

func (msg *Message) Encode() []byte {
	headerBytes := msg.Header.Encode()
	bufLen := len(headerBytes)
	bodyLen := len(msg.Body)
	attachLen := len(msg.Attachment)
	if bodyLen > 0 {
		bufLen += 4 + bodyLen
	} else if attachLen > 0 {
		bufLen += 4
	}
	if attachLen > 0 {
		bufLen += 4 + attachLen
	}
	buf := make([]byte, 0, bufLen)
	buf = append(buf, headerBytes...)
	if bodyLen > 0 {
		index := len(buf)
		buf = append(buf, []byte{0, 0, 0, 0}...)
		binary.BigEndian.PutUint32(buf[index:], uint32(bodyLen))
		buf = append(buf, msg.Body...)
	} else if attachLen > 0 {
		index := len(buf)
		buf = append(buf, []byte{0, 0, 0, 0}...)
		binary.BigEndian.PutUint32(buf[index:], 0)
	}
	if attachLen > 0 {
		index := len(buf)
		buf = append(buf, []byte{0, 0, 0, 0}...)
		binary.BigEndian.PutUint32(buf[index:], uint32(attachLen))
		buf = append(buf, msg.Attachment...)
	}
	return buf
}

func MessageDecode(buf []byte) *Message {
	header, headerSize := HeaderDecode(buf)
	buf = buf[headerSize:]
	var body []byte
	if len(buf) > 4 {
		bodyLen := binary.BigEndian.Uint32(buf)
		if bodyLen > 0 {
			body = make([]byte, bodyLen)
			copy(body, buf[4:4+bodyLen])
		}
		buf = buf[4+bodyLen:]
	}
	var attach []byte
	if len(buf) > 4 {
		attachLen := binary.BigEndian.Uint32(buf)
		if attachLen > 0 {
			if attachLen > uint32(cap(buf)) {
				panic("")
			}
			attach = make([]byte, attachLen)
			copy(attach, buf[4:4+attachLen])
		}
	}
	return &Message{Header: header, Body: body, Attachment: attach}
}

func encode(body interface{}) []byte {
	if body == nil {
		return make([]byte, 0)
	}
	switch body.(type) {
	case []byte:
		return body.([]byte)
	default:
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			panic(errors.New("序列化消息失败" + err.Error()))
		}
		return bodyBytes
	}

}

func MessageValueOf(header *Header, body interface{}, attachment interface{}) *Message {
	bodyBytes := encode(body)
	attachBytes := encode(attachment)
	message := Message{Header: header, Body: bodyBytes, Attachment: attachBytes}
	return &message
}

type MessageHandler interface {
	Receive(msg *Message)
}
