package client

import (
	"fmt"
	"sync"
	"time"
)

type recordItem struct {
	Module  uint16
	Command uint16
	Min     time.Duration
	Max     time.Duration
	Avg     time.Duration
	Times   uint
	sum     time.Duration
}

func (item *recordItem) String() string {
	return fmt.Sprintf("模块[%d,%d],\tmin:%s,\tmax:%s,\tavg:%s,\tcount:%d", int16(item.Module), item.Command, item.Min, item.Max, item.Avg, item.Times)
}

func (item *recordItem) record(duration time.Duration) {
	if item.Min > duration {
		item.Min = duration
	}
	if item.Max < duration {
		item.Max = duration
	}
	item.sum += duration
	item.Times++
	item.Avg = item.sum / time.Duration(item.Times)
}

var QpsRecord = make(map[uint64]*recordItem)
var mutex = sync.Mutex{}

func Record(module uint16, command uint16, duration time.Duration) {
	fmt.Println("收到请求", module, command, duration)
	mutex.Lock()
	key := uint64(module)<<16 + uint64(command)
	preItem := QpsRecord[key]
	if preItem == nil {
		preItem = &recordItem{Module: module, Command: command, Min: 0xFFFFFFFF}
		QpsRecord[key] = preItem
	}
	preItem.record(duration)
	mutex.Unlock()
}
