package messages

import (
	"math"
	"testing"

	"github.com/montanaflynn/stats"
)

func TestRandomMessage(t *testing.T) {
	doCheckSize(10, t)
	doCheckSize(15, t)
	doCheckSize(20, t)
	doCheckSize(25, t)
	doCheckSize(30, t)
	doCheckSize(100, t)
	doCheckSize(110, t)
	doCheckSize(120, t)
	doCheckSize(125, t)
}

func doCheckSize(size int, t *testing.T) {
	message := GenerateRandomMessage(size)
	if len(message) != size {
		t.Errorf("Size of message expected to be %d. Was really %d.", size, len(message))
	}
}

func TestRandomMessageDistribution(t *testing.T) {
	doCheckSizeDistribution(1000, 20, 5.0, t)
}

func doCheckSizeDistribution(num, size int, variance float64, t *testing.T) {
	messages := GenerateRandomMessages(num, size, variance)
	var lengths stats.Float64Data
	for _, msg := range messages {
		lengths = append(lengths, float64(len(msg)))
	}
	if len(lengths) != num {
		t.Errorf("number of lengths (%d) not equal to number of messages (%d)", len(lengths), num)
	}
	mean, err := lengths.Mean()
	if err != nil {
		t.Errorf("Error calculating mean of message sizes")
	}
	// Tolerance should be related to standard error of the mean, which is
	// sqrt(variance/num) assuming uncorrelated samples (which they should be for
	// good PRNGs). 95% confidence is within 2 standard deviations. Target mean
	// is "size"
	std := math.Sqrt(variance / float64(num))
	if math.Abs(mean-float64(size)) > math.Ceil(2.5*std) {
		t.Errorf("Mean (%f) deviates from expected (%d) by too much (allowed %d)",
			mean, size, int(math.Ceil(2.5*std)))
	}

	// Make sure variance is about what I'd expect
	calcvar, err := lengths.Variance()
	if err != nil {
		t.Errorf("Error calculating variance of message sizes")
	}
	if math.Abs(calcvar-variance) > math.Ceil(variance/math.Sqrt(float64(num))) {
		t.Errorf("Variance (%f) deviates from expected (%f) by too much (allowed %f)",
			calcvar, variance, variance/math.Sqrt(float64(num)))
	}
}
