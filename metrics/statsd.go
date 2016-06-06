package metrics

import (
	"fmt"
	"io"
	"time"
)

// StatsDPackMax is the maximum packet size for batches.
// Be carefull not to exceed the network's MTU. See
// https://github.com/etsy/statsd/blob/master/docs/metric_types.md#multi-metric-packets
var StatsDPackMax = 1432

type statsD struct {
	prefix string
	// pool holds buffers to prevent mallocs at runtime.
	pool chan []byte
	// queue holds the pending messages.
	queue chan []byte
}

// NewStatsD returns a new Register with a StatsD implementation.
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
	d := new(statsD)
	{
		size := 1000
		d.queue = make(chan []byte, size)
		d.pool = make(chan []byte, size)
		for i := 0; i < size; i++ {
			d.pool <- make([]byte, 0, 40)
		}
	}

	// write statsD.queue to conn
	go func() {
		buf := make([]byte, 0, StatsDPackMax)
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
			case next = <-d.queue:
			default:
				time.Sleep(time.Millisecond)
				continue
			}

			// flush first when size limit reached
			if len(buf)+1+len(next) > StatsDPackMax {
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
			d.pool <- next[:0]
		}
	}()

	return d
}

func (d *statsD) Seen(key string, n int) {
	fmt.Fprintf(d, "%s%s:%d|c", d.prefix, key, n)
}

func (d *statsD) Took(key string, since time.Time) {
	fmt.Fprintf(d, "%s%s:%d|ms", d.prefix, key, time.Since(since)/time.Millisecond)
}

func (d *statsD) Write(p []byte) (n int, err error) {
	d.queue <- append(<-d.pool, p...)
	return len(p), nil
}

func (d *statsD) KeyPrefix(s string) {
	d.prefix = s
}
