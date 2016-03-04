package metrics

import (
	"fmt"
	"io"
	"time"
)

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
// Batches are limited by the flushInterval span. A flushInterval of 0 implies no buffering.
func NewStatsD(conn io.Writer, flushInterval time.Duration) Register {
	d := new(statsd)
	{
		size := 1000
		d.Queue = make(chan []byte, size)
		d.Pool = make(chan []byte, size)
		for i := 0; i < size; i++ {
			d.Pool <- make([]byte, 0, 40)
		}
	}

	// write statsd.Queue to conn
	go func() {
		buf := make([]byte, 0, statsdSizeLimit)
		var batchStart time.Time
		for {
			// flush on interval timeout
			if len(buf) > 0 && time.Since(batchStart) >= flushInterval {
				conn.Write(buf)
				buf = buf[:0]
				continue
			}

			var next []byte
			select {
			case next = <-d.Queue:
			default:
				time.Sleep(time.Millisecond)
				continue
			}

			// flush first when size limit reached
			if len(buf)+1+len(next) > statsdSizeLimit {
				conn.Write(buf)
				buf = buf[:0]
			}

			if len(buf) != 0 {
				buf = append(buf, '\n')
			} else {
				batchStart = time.Now()
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
	fmt.Fprintf(d, "%s%s:%d|ms", d.Prefix, key, time.Since(since)/time.Millisecond)
}

func (d *statsd) Write(p []byte) (n int, err error) {
	d.Queue <- append(<-d.Pool, p...)
	return len(p), nil
}

func (d *statsd) KeyPrefix(s string) {
	d.Prefix = s
}
