package tunneling

import (
	"context"
	"errors"
	"time"
)

type SourceMock struct {
	out                        chan []byte
	gotMsgs                    []string
	readMsgs                   []string
	failsConsumeBeforeMessages bool
	failsConsumeAfterMessages  bool
	consuming                  bool
	consumingStopDelayMilliSec int
}

func NewSourceMock(
	readMsgs []string,
	failsConsumeBeforeMessages bool, failsConsumeAfterMessages bool,
	consumingStopDelayMilliSec int,
) *SourceMock {
	return &SourceMock{
		out:                        make(chan []byte),
		readMsgs:                   readMsgs,
		gotMsgs:                    make([]string, 0),
		failsConsumeBeforeMessages: failsConsumeBeforeMessages,
		failsConsumeAfterMessages:  failsConsumeAfterMessages,
		consumingStopDelayMilliSec: consumingStopDelayMilliSec,
	}
}

func NewNormalSourceMock(readMsgs []string) *SourceMock {
	return NewSourceMock(readMsgs, false, false, 0)
}

func (s *SourceMock) Consume(ctx context.Context) (err error) {
	if s.failsConsumeBeforeMessages {
		return errors.New("failed before")
	}
	s.consuming = true
	defer func() {
		s.consuming = false
	}()
	for _, msg := range s.readMsgs {
		s.out <- []byte(msg)
	}
	if s.failsConsumeAfterMessages {
		return errors.New("failed after")
	}
	<-ctx.Done()
	time.Sleep(time.Duration(s.consumingStopDelayMilliSec) * time.Millisecond)
	return nil
}

func (s SourceMock) GetReader() chan []byte {
	return s.out
}

func (s *SourceMock) Write(bytes []byte) (err error) {
	s.gotMsgs = append(s.gotMsgs, string(bytes))
	return nil
}

func (s *SourceMock) Connect(ctx context.Context) error {
	return nil
}

func (s *SourceMock) GetUrl() (url string) {
	return ""
}

func (s *SourceMock) gotMessage(msg string) bool {
	for _, m := range s.gotMsgs {
		if m == msg {
			return true
		}
	}
	return false
}
