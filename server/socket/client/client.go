package client

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"
	"xa.com/manager/server/socket/core"
)

var ClientSn int64 = 1
var MessageSn int64 = 1

type messageFuture struct {
	Ch     chan *core.Message
	Create time.Time
}

type Client struct {
	Id      int64
	Connect net.Conn
	Valid   bool
	context.Context
	core.MessageHandler
	Futures  *sync.Map //map[uint64]*messageFuture
	mutex    sync.Mutex
	PushChan chan *core.Message
}

func Connect(address string, timeout time.Duration) (*Client, error) {
	//connectMutex.Lock()
	//defer func() { connectMutex.Unlock() }()
	//pre := Clients[address]
	//if pre != nil && pre.Valid {
	//	return pre, nil
	//}
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		panic(errors.New("连接建立失败" + err.Error()))
	}
	tcpConn, valid := conn.(*net.TCPConn)
	if !valid {
		return nil, errors.New("连接失败")
	}
	err = tcpConn.SetReadBuffer(2 * 1024)
	if err != nil {
		_ = conn.Close()
		return nil, errors.New("设置连接参数失败")
	}
	clientCtx := context.WithValue(context.Background(), "name", address)
	clientId := increaseAndGet(&ClientSn)
	client := &Client{Id: clientId,
		Connect:  conn,
		Valid:    true,
		Context:  clientCtx,
		Futures:  &sync.Map{},
		mutex:    sync.Mutex{},
		PushChan: make(chan *core.Message, 16),
	}
	go client.readMsg(conn)
	return client, nil
}

func (cli *Client) readMsg(conn net.Conn) {
	reqBuf := make([]byte, 128)
	defer func() { _ = conn.Close() }()
	packCache := bytes.NewBuffer(nil)
	for {
		count, err := conn.Read(reqBuf)
		if err != nil {
			fmt.Println("连接从远程关闭", err)
			break
		}
		if count <= 0 {
			fmt.Println("连接从远程关闭")
			break
		}
		packCache.Write(reqBuf[0:count])
		for {
			packLen := uint32(packCache.Len())
			if packLen < 8 {
				break
			}
			pack := packCache.Bytes()
			packIdentity := binary.BigEndian.Uint32(pack)
			if packIdentity != 0xFFFFFFFF {
				fmt.Println("数据包非法", packIdentity)
				packCache.Reset()
				break
			}
			bodyLen := binary.BigEndian.Uint32(pack[4:])
			curPackSize := 4 + 4 + bodyLen
			// 数据包还不完整
			if packLen < curPackSize {
				break
			}
			pack = packCache.Next(int(curPackSize))
			checkSumIndex := 8 + bodyLen - 4
			message := core.MessageDecode(pack[8:checkSumIndex])

			checkSum := binary.BigEndian.Uint32(pack[checkSumIndex:])

			calCheckSum := CheckSum(pack[8:checkSumIndex])
			if int32(checkSum) != calCheckSum {
				fmt.Println("数据包校验不合格")
			}
			cli.Receive(message)
		}
		if packCache.Len() == 0 {
			packCache.Reset()
		}
	}

}

func CheckSum(array []byte) int32 {
	var hashcode int64 = 0

	length := len(array)
	for i := 0; i < length; i++ {
		hashcode = hashcode<<7 ^ int64(int8(array[i]))
	}
	return int32(hashcode)
}

func (c *Client) Receive(msg *core.Message) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("收取数据包异常", msg.Header, err)
		}
	}()
	// 推送数据包
	if msg.Header.IsRequest() {
		fmt.Println(c.Id, "收到推送数据包", msg.ToString())
		return
	}
	sn := msg.Header.Sn
	c.mutex.Lock()
	defer c.mutex.Unlock()
	future, _ := c.Futures.Load(sn)
	if future == nil {
		fmt.Println(c.Id, "收到数据包，未发现sn，可能已超时", msg.ToString())
		return
	}
	msgFuture := future.(*messageFuture)
	Record(msg.Module, msg.Command, time.Now().Sub(msgFuture.Create))
	c.Futures.Delete(sn)
	msgFuture.Ch <- msg
	close(msgFuture.Ch)
}

func (c *Client) Write(module, command uint16, value interface{}) (<-chan *core.Message, error) {
	sn := uint64(increaseAndGet(&MessageSn))
	header := &core.Header{Format: 1, State: 0, Sn: sn, Session: 0, Module: module, Command: command}
	msg := core.MessageValueOf(header, value, nil)
	buf := msg.Encode()
	totalSize := 4 + 4 + len(buf) + 4
	pack := make([]byte, 0, totalSize)

	b4 := make([]byte, 4)
	// 添加消息头
	pack = append(pack, b4...)
	binary.BigEndian.PutUint32(pack, 0xFFFFFFFF)
	// 消息长度
	pack = append(pack, b4...)
	binary.BigEndian.PutUint32(pack[4:], 4+uint32(len(buf)))
	// 消息体,保证不会扩容
	pack = append(pack, buf...)
	index := len(pack)
	// checksum
	checkSum := uint32(CheckSum(buf))
	pack = append(pack, b4...)
	binary.BigEndian.PutUint32(pack[index:], checkSum)
	if c.Connect == nil {
		fmt.Println("connect为空")
		return nil, errors.New("连接为空")
	}
	count, err := c.Connect.Write(pack)
	if err != nil {
		fmt.Println("写入数据错误", count, err)
		return nil, errors.New("写入数据错误")
	}
	//fmt.Println(c.Id, "成功发送数据量[", totalSize, "]已发送[", count, "]")
	future := make(chan *core.Message)
	msgFuture := &messageFuture{Ch: future, Create: time.Now()}
	c.Futures.Store(sn, msgFuture)
	time.AfterFunc(5*time.Second, func() {
		c.mutex.Lock()
		defer c.mutex.Unlock()
		preFu, _ := c.Futures.Load(sn)
		if preFu != nil {
			c.Futures.Delete(sn)
			close(future)
		}
	})
	return future, nil
}

func increaseAndGet(value *int64) int64 {
	for {
		oldValue := *value
		newValue := oldValue + 1
		suc := atomic.CompareAndSwapInt64(value, oldValue, newValue)
		if suc {
			return newValue
		}
	}
}
