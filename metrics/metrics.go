package metrics

import (
	"fmt"
	"io"
	"time"
)

type Register interface {
	// Seen increments the counter with n.
	Seen(key string, n int)

	// Took adds a timing from since to now.
	Took(key string, since time.Time)

	// KeyPrefix defines a prefix applied to all keys.
	KeyPrefix(string)
}

type dummy struct {}

func NewDummy() Register {
	return dummy{}
}

func (d dummy) Seen(key string, n int) {}
func (d dummy) Took(key string, since time.Time) {}
func (d dummy) KeyPrefix(s string) {}

const statsdSizeLimit = 65536 - 8 - 20 // 8-byte UDP header, 20-byte IP header

type statsd struct {
	Prefix string
	// Pool holds buffers to prevent mallocs at runtime.
	Pool chan []byte
	// Queue holds the pending messages.
	Queue chan []byte
}

// NewStatsd returns a new Register with a StatsD implementation.
// Generally speaking, you want to connect over UDP.
//
//	conn, err := net.DialTimeout("udp", "localhost:8125", 4*time.Second)
//	if err != nil {
//		...
//	}
//	stats := metrics.NewStatsD(conn, time.Second)
//
// Batches are send at least once every flushInterval. A flushInterval of 0 implies no buffering.
func NewStatsD(conn io.Writer, flushInterval time.Duration) Register {
	d := new(statsd)
	{
		size := 1000
		d.Queue = make(chan []byte, size)
		d.Pool = make(chan []byte, size)
		for i := 0; i < size; i++ {
			d.Pool <- make([]byte, 0, 80)
		}
	}

	// write statsd.Queue to conn
	go func() {
		buf := make([]byte, statsdSizeLimit)
		lastFlush := time.Now()
		for {
			// flush on interval timeout
			if len(buf) > 0 && time.Since(lastFlush) >= flushInterval {
				conn.Write(buf)
				lastFlush = time.Now()
				buf = buf[:0]
				continue
			}

			// pause on no data
			if len(d.Queue) == 0 {
				time.Sleep(time.Millisecond)
				continue
			}

			next := <- d.Queue

			// flush first when size limit reached
			if len(buf) + 1 + len(next) > statsdSizeLimit {
				conn.Write(buf)
				lastFlush = time.Now()
				buf = buf[:0]
			}

			if len(buf) != 0 {
				buf = append(buf, '\n')
			}
			buf = append(buf, next...)

			// reuse buffer
			d.Pool <- next[:0]
		}
	}()

	return d
}

func (d *statsd) Seen(key string, n int) {
	fmt.Fprintf(d, "%s%s:%d|c", d.Prefix, key, n)
}

func (d *statsd) Took(key string, since time.Time) {
	fmt.Fprintf(d, "%s%s:%d|ms", d.Prefix, key, time.Since(since) / time.Millisecond)
}

func (d *statsd) Write(p []byte) (n int, err error) {
	d.Queue <- append(<- d.Pool, p...)
	return len(p), nil
}

func (d *statsd) KeyPrefix(s string) {
	d.Prefix = s
}
