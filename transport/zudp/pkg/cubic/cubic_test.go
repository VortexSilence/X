// cubic_test.go
package cubic

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCubicSlowStart(t *testing.T) {
	c := NewCubic()

	// Slow Start
	for i := 0; i < 10; i++ {
		c.OnPacketSent(1460)
		c.OnPacketAck(1460, 50*time.Millisecond)
	}

	assert.True(t, c.WindowSize() > 10*1460, "Window should grow exponentially in slow start")
	assert.Equal(t, uint32(math.MaxUint32), c.ssthresh, "ssthresh should remain max in slow start")
}

func TestCubicCongestionAvoidance(t *testing.T) {
	c := NewCubic()

	// Trigger packet loss to exit slow start
	c.OnPacketSent(1460)
	c.OnPacketLost(1460)

	initialWindow := c.WindowSize()

	// Congestion Avoidance
	for i := 0; i < 20; i++ {
		c.OnPacketSent(1460)
		c.OnPacketAck(1460, 50*time.Millisecond)
	}

	assert.True(t, c.WindowSize() > initialWindow, "Window should grow in congestion avoidance")
	assert.True(t, c.WindowSize() < initialWindow+20*1460, "Growth should be slower than slow start")
}

func TestCubicPacketLoss(t *testing.T) {
	c := NewCubic()

	// Fill window
	for i := 0; i < 10; i++ {
		c.OnPacketSent(1460)
		c.OnPacketAck(1460, 50*time.Millisecond)
	}

	windowBeforeLoss := c.WindowSize()
	c.OnPacketSent(1460)
	c.OnPacketLost(1460)
	//TODO check this plz
	fmt.Println(uint32(float64(windowBeforeLoss) * cubicBeta))
	fmt.Println(c.WindowSize())
	assert.True(t, c.WindowSize() <= uint32(float64(windowBeforeLoss)*cubicBeta),
		"Window should decrease after loss")
	assert.True(t, c.ssthresh < windowBeforeLoss,
		"ssthresh should be reduced after loss")
}

func TestRTTCalculation(t *testing.T) {
	c := NewCubic()

	c.OnPacketSent(1460)
	c.OnPacketAck(1460, 100*time.Millisecond)

	assert.Equal(t, 100*time.Millisecond, c.minRTT, "Should update min RTT")

	c.OnPacketSent(1460)
	c.OnPacketAck(1460, 80*time.Millisecond)

	assert.Equal(t, 80*time.Millisecond, c.minRTT, "Should update min RTT with lower value")
}
