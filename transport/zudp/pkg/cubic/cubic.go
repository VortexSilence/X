package cubic

import (
	"math"
	"sync"
	"time"
)

const (
	initialWindowSize    = 16 * 1024        // 16KB
	defaultMaxWindowSize = 10 * 1024 * 1024 // 10MB
	minWindowSize        = 2 * 1024         // 2KB
	cubicBeta            = 0.7              // فاکتور کاهش پنجره
	cubicC               = 0.4              // ثابت CUBIC
	cubicFastConverge    = 0.85             // فاکتور همگرایی سریع
)

type Cubic struct {
	mu                sync.Mutex
	windowSize        uint32  // اندازه پنجره کنونی (بایت)
	ssthresh          uint32  // آستانه slow start
	lastMaxWindow     float64 // آخرین حداکثر پنجره قبل از کاهش
	windowAtReduction float64 // اندازه پنجره در زمان کاهش
	k                 float64 // پارامتر زمانی CUBIC
	epochStart        time.Time
	minRTT            time.Duration
	smoothedRTT       time.Duration
	rttVar            time.Duration
	bytesInFlight     uint32
	packetsAcked      uint32
	packetsLost       uint32
	maxWindowSize     uint32
	lastUpdate        time.Time
}

func NewCubic() *Cubic {
	return &Cubic{
		windowSize:    initialWindowSize,
		ssthresh:      math.MaxUint32,
		maxWindowSize: defaultMaxWindowSize,
	}
}

func (c *Cubic) WindowSize() uint32 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.windowSize
}

func (c *Cubic) SetMaxWindow(size uint32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.maxWindowSize = size
	if c.windowSize > c.maxWindowSize {
		c.windowSize = c.maxWindowSize
	}
}

func (c *Cubic) OnPacketSent(bytesSent uint32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.bytesInFlight += bytesSent
}

func (c *Cubic) OnPacketAck(ackedBytes uint32, rtt time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.updateRTT(rtt)
	c.packetsAcked++
	c.bytesInFlight -= ackedBytes

	if c.windowSize < c.ssthresh {
		// Slow Start
		c.windowSize += ackedBytes
		if c.windowSize > c.maxWindowSize {
			c.windowSize = c.maxWindowSize
		}
	} else {
		// Congestion Avoidance با CUBIC
		c.congestionAvoidance(ackedBytes, rtt)
	}
}

func (c *Cubic) OnPacketLost(lostBytes uint32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.packetsLost++
	c.bytesInFlight -= lostBytes

	// ذخیره اندازه پنجره قبل از کاهش
	c.windowAtReduction = float64(c.windowSize)
	if c.lastMaxWindow < c.windowAtReduction {
		c.lastMaxWindow = c.windowAtReduction
	}

	// محاسبه آستانه جدید
	c.ssthresh = uint32(math.Max(
		float64(c.windowSize)*cubicFastConverge,
		float64(c.windowSize)*cubicBeta,
	))

	// کاهش پنجره
	c.windowSize = uint32(math.Max(
		float64(minWindowSize),
		float64(c.windowSize)*cubicBeta,
	))

	// محاسبه K جدید
	c.k = math.Cbrt((c.lastMaxWindow * (1 - cubicBeta)) / cubicC)
	c.epochStart = time.Time{} // Reset epoch
}

func (c *Cubic) congestionAvoidance(ackedBytes uint32, rtt time.Duration) {
	if c.epochStart.IsZero() {
		c.epochStart = time.Now()
	}

	t := time.Since(c.epochStart).Seconds()
	target := c.cubicWindow(t)

	// TCP-Friendly check
	tcpTarget := c.tcpFriendlyWindow(t)
	if tcpTarget < target {
		target = tcpTarget
	}

	// افزایش پنجره
	if target > float64(c.windowSize) {
		c.windowSize = uint32(math.Min(float64(c.maxWindowSize), target))
	} else {
		c.windowSize = uint32(math.Max(float64(c.windowSize), target))
	}
}

func (c *Cubic) cubicWindow(t float64) float64 {
	return c.lastMaxWindow*cubicC + cubicC*math.Pow(t-c.k, 3)
}

func (c *Cubic) tcpFriendlyWindow(t float64) float64 {
	if c.smoothedRTT == 0 {
		return 0
	}
	return c.lastMaxWindow*(1-cubicBeta) + 3*((1-cubicBeta)/(1+cubicBeta))*(t/c.smoothedRTT.Seconds())
}

func (c *Cubic) updateRTT(rtt time.Duration) {
	if c.minRTT == 0 || rtt < c.minRTT {
		c.minRTT = rtt
	}

	const (
		alpha = 0.125 // فیلتر برای RTT میانگین
		beta  = 0.25  // فیلتر برای تغییرات RTT
	)

	if c.smoothedRTT == 0 {
		c.smoothedRTT = rtt
		c.rttVar = rtt / 2
	} else {
		delta := c.smoothedRTT - rtt
		if delta < 0 {
			delta = -delta
		}
		c.rttVar = time.Duration((1-beta)*float64(c.rttVar) + beta*float64(delta))
		c.smoothedRTT = time.Duration((1-alpha)*float64(c.smoothedRTT) + alpha*float64(rtt))
	}

	c.lastUpdate = time.Now()
}

func (c *Cubic) AvailableWindow() uint32 {
	c.mu.Lock()
	defer c.mu.Unlock()

	// محاسبه پنجره قابل استفاده
	if c.bytesInFlight >= c.windowSize {
		return 0
	}
	return c.windowSize - c.bytesInFlight
}
