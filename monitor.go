package monitor

import (
	"io"
	"math"
	"sync"
	"time"
)

// clockRate is the resolution and precision of clock().
const clockRate = 10 * time.Millisecond

// czero is the process start time rounded down to the nearest clockRate
// increment.
var czero = time.Duration(time.Now().UnixNano()) / clockRate * clockRate

type Reader struct {
	io.Reader
	*Monitor
}

func NewReader(monitor *Monitor, r io.Reader) io.ReadCloser {
	return &Reader{r, monitor}
}

func (r *Reader) Read(p []byte) (n int, err error) {
	n, err = r.IO(r.Reader.Read(p))
	return
}

func (r *Reader) Close() error {
	defer r.Done()
	if c, ok := r.Reader.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

func (r *Reader) Seek(offset int64, whence int) (int64, error) {
	if c, ok := r.Reader.(io.Seeker); ok {
		return c.Seek(offset, whence)
	}
	return 0, nil
}

// Monitor holds information of IO flowing through it
type Monitor struct {
	Start   time.Duration // transfer start time
	Samples uint64        // total number of samples
	Bytes   uint64        // total number of bytes transferred
	Peak    float64       // peak transfer rate of all samples
	Rsample float64       // most recently taken sample
	EMA     float64
	Sbytes  uint64        // number of bytes transferred since Slast
	RWindow float64       // most recent window
	Slast   time.Duration // most recent sample time
	Srate   time.Duration // sampling rate
	Tlast   time.Duration // time of the most recent transfer

	mu sync.Mutex
}

func NewMonitor(sample, windowSize time.Duration) *Monitor {
	if sample = clockRound(sample); sample <= 0 {
		sample = 5 * clockRate
	}
	if windowSize <= 0 {
		windowSize = 1 * time.Second
	}

	now := clock()
	return &Monitor{
		Start:   now,
		RWindow: windowSize.Seconds(),
		Slast:   now,
		Srate:   sample,
		Tlast:   now,
	}
}

func (m *Monitor) Done() int64 {
	m.mu.Lock()
	now := m.update(0)
	if m.Sbytes > 0 {
		m.reset(now)
	}
	m.Tlast = 0
	n := m.Bytes
	m.mu.Unlock()
	return int64(n)
}

// update accumulates the transferred byte count for the current sample
func (m *Monitor) update(n int) (now time.Duration) {
	if now = clock(); n > 0 {
		m.Tlast = now
	}

	m.Sbytes += uint64(n)
	if sTime := now - m.Slast; sTime >= m.Srate {
		t := sTime.Seconds()
		if m.Rsample = float64(m.Sbytes) / t; m.Rsample > m.Peak {
			m.Peak = m.Rsample
		}

		if m.Samples > 0 {
			w := math.Round(-t / m.RWindow)

			if w < 0 {
				m.EMA = 0
			} else {
				m.EMA = m.Rsample + w*(m.EMA-m.Rsample)
			}
		} else {
			m.EMA = m.Rsample
		}

		m.reset(now)
	}

	return
}

func (m *Monitor) reset(sample time.Duration) {
	m.Bytes += m.Sbytes
	m.Samples++
	m.Sbytes = 0
	m.Slast = sample
}

func (m *Monitor) Update(n int) int {
	m.mu.Lock()
	m.update(n)
	m.mu.Unlock()
	return n
}

func (m *Monitor) IO(n int, err error) (int, error) {
	return m.Update(n), err
}

// returns current status of the monitor
func (m *Monitor) Status() Status {
	m.mu.Lock()
	now := m.update(0)
	s := Status{
		Start:       clockToTime(m.Start),
		Duration:    m.Slast - m.Start,
		Idle:        now - m.Tlast,
		Bytes:       m.Bytes,
		Samples:     m.Samples,
		PeakRate:    round(m.Peak),
		CurrentRate: round(m.EMA),
	}

	if s.Duration > 0 {
		avg := float64(s.Bytes) / s.Duration.Seconds()
		s.AvgRate = round(avg)
	}

	m.mu.Unlock()

	return s
}

// Status represents the current Monitor status
type Status struct {
	Start       time.Time     // transfer start time
	Duration    time.Duration // time period of the status
	Idle        time.Duration // time since last bytes
	Bytes       uint64        // total number of bytes transferred
	Samples     uint64        // total number of samples taken
	Rate        uint64        // current transfer rate
	PeakRate    uint64        // maximum transfer rate
	CurrentRate uint64        // current transfer rate
	AvgRate     uint64        // average transfer rate of the duration
}

// clock returns a low resolution timestamp relative to the process start time.
func clock() time.Duration {
	return time.Duration(time.Now().UnixNano())/clockRate*clockRate - czero
}

func clockRound(d time.Duration) time.Duration {
	return (d + clockRate>>1) / clockRate * clockRate
}

func round(x float64) uint64 {
	if _, frac := math.Modf(x); frac >= 0.5 {
		return uint64(math.Ceil(x))
	}
	return uint64(math.Floor(x))
}

func clockToTime(c time.Duration) time.Time {
	return time.Unix(0, int64(czero+c))
}
