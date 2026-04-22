package core

import (
	"math/big"
	"testing"
)

func TestIsSequenceTooFarAhead(t *testing.T) {
	c := &Core{}
	threshold := int64(1)

	tests := []struct {
		viewSeq  int64
		currSeq  int64
		expected bool
	}{
		{100, 100, false}, // same
		{101, 100, false}, // next (diff=1) - SHOULD NOT BE TOO FAR
		{102, 100, true},  // diff=2 - TOO FAR
	}

	for _, tt := range tests {
		_, tooFar := c.isSequenceTooFarAhead(big.NewInt(tt.viewSeq), big.NewInt(tt.currSeq), threshold)
		if tooFar != tt.expected {
			t.Errorf("isSequenceTooFarAhead(%d, %d, %d) = %v; want %v", tt.viewSeq, tt.currSeq, threshold, tooFar, tt.expected)
		}
	}
}

func TestIsRoundTooFarAhead(t *testing.T) {
	c := &Core{}
	threshold := int64(10)

	tests := []struct {
		viewRound int64
		currRound int64
		expected  bool
	}{
		{0, 0, false},
		{5, 0, false},
		{10, 0, false}, // diff=10 - SHOULD NOT BE TOO FAR
		{11, 0, true},  // diff=11 - TOO FAR
	}

	for _, tt := range tests {
		_, tooFar := c.isRoundTooFarAhead(big.NewInt(tt.viewRound), big.NewInt(tt.currRound), threshold)
		if tooFar != tt.expected {
			t.Errorf("isRoundTooFarAhead(%d, %d, %d) = %v; want %v", tt.viewRound, tt.currRound, threshold, tooFar, tt.expected)
		}
	}
}
