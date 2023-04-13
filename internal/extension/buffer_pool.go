package extension

import "go.uber.org/zap/buffer"

var (
	_bufferPool = buffer.NewPool()
	// getBuffer retrieves a buffer from the pool, creating one if necessary.
	getBuffer = _bufferPool.Get
)
