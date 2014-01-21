package aws

import (
	"time"
)

/**
 * AttemptStrategy represents a strategy for waiting for an action
 * to complete successfully. This is an internal type used by the
 * implementation of other goamz packages.
 */
type AttemptStrategy struct {
	Total time.Duration // total duration of attempt.
	Delay time.Duration // interval between each try in the burst.
	Min   int           // minimum number of retries; overrides Total
}

type Attempt struct {
	strategy AttemptStrategy
	last     time.Time
	end      time.Time
	force    bool
	count    int
}

/**
 * Start begins a new sequence of attempts for the given strategy.
 */
func (self AttemptStrategy) Start() *Attempt {
	now := time.Now()

	return &Attempt{
		strategy: self,
		last:     now,
		end:      now.Add(self.Total),
		force:    true,
	}
}

/**
 * Next waits until it is time to perform the next attempt or returns
 * false if it is time to stop trying.
 */
func (self *Attempt) Next() bool {
	now := time.Now()
	sleep := self.nextSleep(now)

	if !self.force && !now.Add(sleep).Before(self.end) && self.strategy.Min <= self.count {
		return false
	}

	self.force = false

	if sleep > 0 && self.count > 0 {
		time.Sleep(sleep)
		now = time.Now()
	}

	self.count++
	self.last = now

	return true
}

func (self *Attempt) nextSleep(now time.Time) time.Duration {
	sleep := self.strategy.Delay - now.Sub(self.last)

	if sleep < 0 {
		return 0
	}

	return sleep
}

/** 
 * HasNext returns whether another attempt will be made if the current
 * one fails. If it returns true, the following call to Next is
 * guaranteed to return true.
 */
func (self *Attempt) HasNext() bool {
	if self.force || self.strategy.Min > self.count {
		return true
	}

	now := time.Now()

	if now.Add(self.nextSleep(now)).Before(self.end) {
		self.force = true
		return true
	}

	return false
}
