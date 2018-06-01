package protocol

import (
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"github.com/zuiwuchang/king-go/container/queue"
	"io"
	"sync"
)

var errSendQuqueExit = errors.New("SendQuque already quit")

// Hash .
func Hash(key, val string) (str string, e error) {
	val = key + val
	sha := sha512.New()
	_, e = sha.Write([]byte([]byte(val)))
	if e != nil {
		return
	}
	str = hex.EncodeToString(sha.Sum(nil))
	return
}

// SendQuque tcp send 隊列
type SendQuque struct {
	w          io.Writer
	buffer     []byte
	bufferSize int
	quque      queue.IQueue
	cond       *sync.Cond

	wait bool
	exit bool
}

// NewSendQuque .
func NewSendQuque(w io.Writer, buffer int) (q *SendQuque, e error) {
	var qm queue.IQueue
	qm, e = queue.NewStepQueue(20, 0)
	if e != nil {
		return
	}
	q = &SendQuque{
		w:      w,
		buffer: make([]byte, buffer),
		quque:  qm,
		cond:   sync.NewCond(&sync.Mutex{}),
	}
	return
}

// Run 運行 發送 goroutine
func (q *SendQuque) Run() (e error) {
	cond := q.cond
	for {
		cond.L.Lock()
		// 是否需要 退出
		if q.exit {
			cond.L.Unlock()
			break
		}

		// 獲取 待send數據
		if q.quque.Len() == 0 {
			// 設置標記 需要 正在等待
			q.wait = true
			cond.Wait()

			// 重置 等待 狀態
			q.wait = false
		}
		// 拷貝數據到 緩衝區
		q.copyToBuffer()
		cond.L.Unlock()

		// 發送 數據
		_, e = q.w.Write(q.buffer[:q.bufferSize])
		if e != nil {
			break
		}
		q.bufferSize = 0
	}
	return
}
func (q *SendQuque) copyToBuffer() {
	quque := q.quque

	for q.quque.Len() != 0 {
		ele, _ := quque.PopFront()
		msg := ele.(Message)
		if len(msg)+q.bufferSize > len(q.buffer) {
			// 超過 最大緩存 結束 copy
			quque.PushFront(msg)
			break
		}
		q.bufferSize += copy(q.buffer[q.bufferSize:], msg)
	}
}

// Exit 通知 Run 退出
func (q *SendQuque) Exit() {
	cond := q.cond
	cond.L.Lock()
	//設置 退出標記
	q.exit = true
	if q.wait {
		//存在 等待 goroutine 喚醒 她
		cond.Signal()
	}
	cond.L.Unlock()
}

// Send 發送 數據
func (q *SendQuque) Send(msg ...Message) (e error) {
	cond := q.cond
	cond.L.Lock()
	if q.exit {
		cond.L.Unlock()
		e = errSendQuqueExit
		return
	}
	//增加 工作
	for _, m := range msg {
		q.quque.PushBack(m)
	}
	if q.wait {
		//存在 等待 goroutine 喚醒 她
		cond.Signal()
	}
	cond.L.Unlock()
	return
}
