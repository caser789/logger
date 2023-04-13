package writer

import (
	"io"
	"sync"
	"time"
)

func NewDoubleBufWriterSize(w io.Writer, size int) BufferedWriter {
	if size <= 0 {
		size = defaultBufSize
	}
	wr := &doubleBufferWriter{
		master: make([]byte, size),
		slave:  make([]byte, size),
		wr:     w,
		size:   size,
		done:   make(chan struct{}),
		sync:   make(chan struct{}, 1),
		cond:   sync.NewCond(&sync.Mutex{}),
	}
	go wr.flushPeriodically()
	return wr
}

/*
master  -- 主缓存，用于写数据
slave   -- 副缓存，用于同步数据
n       -- 当前主缓存写位置
p、q     -- 副缓存同步数据的起始和结束位置
size     -- 缓存大小
wr       -- 底层写接口
*/
type doubleBufferWriter struct {
	master []byte
	slave  []byte
	n      int
	p      int
	q      int
	size   int
	err    error
	wr     io.Writer
	done   chan struct{}
	sync   chan struct{}
	cond   *sync.Cond
	guard  sync.Mutex
	closed bool
}

// 同步副缓存。因为Write返回的可能小于master的长度，因而用p、q分别标识写入的起始和结束位置，当p、q相等时表示同步完成
func (b *doubleBufferWriter) flush() error {
	b.cond.L.Lock()
	defer b.cond.L.Unlock()

	if b.p == b.q {
		return nil
	}
	n, err := b.wr.Write(b.slave[b.p:b.q])
	if n+b.p < b.q && err == nil {
		err = io.ErrShortWrite
	}
	b.p += n
	if err != nil {
		return err
	}
	b.cond.Signal()
	return nil
}

// 先同步副缓存，若主缓存也有数据，主副交换再进行同步
func (b *doubleBufferWriter) flushAll() error {
	if err := b.flush(); err != nil {
		return err
	}
	if b.buffered() {
		b.swap()
		if err := b.flush(); err != nil {
			return err
		}
	}
	return nil
}

func (b *doubleBufferWriter) Flush() error {
	if b.closed {
		return nil
	}
	if err := b.flushAll(); err != nil {
		return err
	}
	close(b.done)
	b.closed = true
	return b.err
}

func (b *doubleBufferWriter) buffered() bool {
	b.guard.Lock()
	ok := b.n > 0
	b.guard.Unlock()
	return ok
}

func (b *doubleBufferWriter) swap() {
	b.cond.L.Lock()
	b.guard.Lock()
	b.master, b.slave = b.slave, b.master
	b.p = 0
	b.q = b.n
	b.n = 0
	b.guard.Unlock()
	b.cond.L.Unlock()
}

// 将p复制到master中，若master满了，判断slave是否已同步完，若是，交换主副缓存，重复上述过程直到p已全部写入。若未同步完，写入阻塞。
func (b *doubleBufferWriter) Write(p []byte) (nn int, err error) {
	for len(p) > 0 {
		if b.err != nil {
			return nn, b.err
		}

		b.guard.Lock()
		n := copy(b.master[b.n:], p)
		b.n += n
		nn += n
		p = p[n:]
		full := b.n == b.size
		b.guard.Unlock()

		if full {
			b.cond.L.Lock()
			synced := b.p == b.q
			b.cond.L.Unlock()

			if !synced {
				b.cond.L.Lock()
				for b.p != b.q {
					b.cond.Wait()
				}
				b.cond.L.Unlock()
			}
			b.swap()
			b.sync <- struct{}{}
		}
	}
	return nn, nil
}

func (b *doubleBufferWriter) flushPeriodically() {
	ticker := time.NewTicker(defaultFlushPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
		case <-b.sync:
		case <-b.done:
			return
		}
		if err := b.flushAll(); err != nil {
			b.err = err
		}
	}
}
