package rkprom

import (
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/stretchr/testify/assert"
	"go.uber.org/atomic"
	"io/ioutil"
	"net/http"
	"runtime"
	"strings"
	"testing"
	"time"
)

var (
	intervalMS       = 200 * time.Millisecond
	remoteAddr       = "localhost:1608"
	jobName          = "rk-prom-job"
	basicAuth        = "user:pass"
	zapLoggerEntry   = rkentry.NoopZapLoggerEntry()
	eventLoggerEntry = rkentry.NoopEventLoggerEntry()
)

type HTTPDoerMock struct {
	called *atomic.Bool
}

func (mock *HTTPDoerMock) Do(*http.Request) (*http.Response, error) {
	mock.called.CAS(false, true)
	return &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(strings.NewReader("mock response")),
	}, nil
}

func TestNewPushGatewayPublisher_WithNegativeIntervalMS(t *testing.T) {
	negative := -1 * time.Microsecond

	pusher, err := NewPushGatewayPusher(
		WithIntervalMSPusher(negative),
		WithRemoteAddressPusher(remoteAddr),
		WithJobNamePusher(jobName),
		WithBasicAuthPusher(basicAuth),
		WithZapLoggerEntryPusher(zapLoggerEntry),
		WithEventLoggerEntryPusher(eventLoggerEntry))

	assert.Nil(t, pusher, "pusher should be nil")
	assert.NotNil(t, err, "error should not be nil")
}

func TestNewPushGatewayPublisher_WithEmptyURL(t *testing.T) {
	pusher, err := NewPushGatewayPusher(
		WithIntervalMSPusher(intervalMS),
		WithRemoteAddressPusher(""),
		WithJobNamePusher(jobName),
		WithBasicAuthPusher(basicAuth),
		WithZapLoggerEntryPusher(zapLoggerEntry),
		WithEventLoggerEntryPusher(eventLoggerEntry))

	assert.Nil(t, pusher, "pusher should be nil")
	assert.NotNil(t, err, "error should not be nil")
}

func TestNewPushGatewayPublisher_WithEmptyJobName(t *testing.T) {
	pusher, err := NewPushGatewayPusher(
		WithIntervalMSPusher(intervalMS),
		WithRemoteAddressPusher(remoteAddr),
		WithJobNamePusher(""),
		WithBasicAuthPusher(basicAuth),
		WithZapLoggerEntryPusher(zapLoggerEntry),
		WithEventLoggerEntryPusher(eventLoggerEntry))

	assert.Nil(t, pusher, "pusher should be nil")
	assert.NotNil(t, err, "error should not be nil")
}

func TestNewPushGatewayPublisher_WithNilLZapLoggerEntry(t *testing.T) {
	pusher, err := NewPushGatewayPusher(
		WithIntervalMSPusher(intervalMS),
		WithRemoteAddressPusher(remoteAddr),
		WithJobNamePusher(jobName),
		WithBasicAuthPusher(basicAuth),
		WithZapLoggerEntryPusher(nil),
		WithEventLoggerEntryPusher(eventLoggerEntry))

	assert.NotNil(t, pusher, "pusher should not be nil")
	assert.Nil(t, err, "error should be nil")
}

func TestNewPushGatewayPublisher_WithNilLEventLoggerEntry(t *testing.T) {
	pusher, err := NewPushGatewayPusher(
		WithIntervalMSPusher(intervalMS),
		WithRemoteAddressPusher(remoteAddr),
		WithJobNamePusher(jobName),
		WithBasicAuthPusher(basicAuth),
		WithZapLoggerEntryPusher(zapLoggerEntry),
		WithEventLoggerEntryPusher(nil))

	assert.NotNil(t, pusher, "pusher should not be nil")
	assert.Nil(t, err, "error should be nil")
}

func TestNewPushGatewayPublisher_WithEmptyBasicAuth(t *testing.T) {
	pusher, err := NewPushGatewayPusher(
		WithIntervalMSPusher(intervalMS),
		WithRemoteAddressPusher(remoteAddr),
		WithJobNamePusher(jobName),
		WithBasicAuthPusher(""),
		WithZapLoggerEntryPusher(zapLoggerEntry),
		WithEventLoggerEntryPusher(eventLoggerEntry))

	assert.NotNil(t, pusher, "pusher should not be nil")
	assert.Nil(t, err, "error should be nil")
}

func TestNewPushGatewayPublisher_WithInvalidBasicAuth(t *testing.T) {
	pusher, err := NewPushGatewayPusher(
		WithIntervalMSPusher(intervalMS),
		WithRemoteAddressPusher(remoteAddr),
		WithJobNamePusher(jobName),
		WithBasicAuthPusher("user:pass:pass"),
		WithZapLoggerEntryPusher(zapLoggerEntry),
		WithEventLoggerEntryPusher(eventLoggerEntry))

	assert.NotNil(t, pusher, "pusher should not be nil")
	assert.Nil(t, err, "error should be nil")
}

func TestNewPushGatewayPublisher_HappyCase(t *testing.T) {
	pusher, err := NewPushGatewayPusher(
		WithIntervalMSPusher(intervalMS),
		WithRemoteAddressPusher(remoteAddr),
		WithJobNamePusher(jobName),
		WithBasicAuthPusher(basicAuth),
		WithZapLoggerEntryPusher(zapLoggerEntry),
		WithEventLoggerEntryPusher(eventLoggerEntry))

	assert.NotNil(t, pusher, "pusher should not be nil")
	assert.Nil(t, err, "error should be nil")

	assert.Equal(t, intervalMS, pusher.IntervalMS)
	assert.NotNil(t, pusher.Pusher, "pusher should not be nil")
	assert.Equal(t, jobName, pusher.JobName)
	assert.Equal(t, basicAuth, pusher.Credential)
	assert.NotNil(t, pusher.lock, "lock should not be nil")
	assert.False(t, pusher.Running.Load(), "isRunning should be false")
}

func TestPushGatewayPusher_Start_WithDuplicateStartCalls(t *testing.T) {
	pusher, err := NewPushGatewayPusher(
		WithIntervalMSPusher(intervalMS),
		WithRemoteAddressPusher(remoteAddr),
		WithJobNamePusher(jobName),
		WithBasicAuthPusher(basicAuth),
		WithZapLoggerEntryPusher(zapLoggerEntry),
		WithEventLoggerEntryPusher(eventLoggerEntry))

	assert.NotNil(t, pusher, "pusher should not be nil")
	assert.Nil(t, err, "error should be nil")

	// we expect a new go routines generated
	prev := runtime.NumGoroutine()

	// start at first time
	pusher.Start()
	assert.True(t, pusher.IsRunning())
	assert.Equal(t, prev+1, runtime.NumGoroutine())

	// call Start() again
	pusher.Start()
	// number of goroutines should be the same
	assert.Equal(t, prev+1, runtime.NumGoroutine())

	// call Stop()
	pusher.Stop()
}

func TestPushGatewayPusher_Start_HappyCase(t *testing.T) {
	pusher, err := NewPushGatewayPusher(
		WithIntervalMSPusher(intervalMS),
		WithRemoteAddressPusher(remoteAddr),
		WithJobNamePusher(jobName),
		WithBasicAuthPusher(basicAuth),
		WithZapLoggerEntryPusher(zapLoggerEntry),
		WithEventLoggerEntryPusher(eventLoggerEntry))

	// mock a pusher for tracing
	pusher.Pusher = pusher.Pusher.Client(&HTTPDoerMock{
		called: atomic.NewBool(false),
	})

	assert.NotNil(t, pusher, "pusher should not be nil")
	assert.Nil(t, err, "error should be nil")

	// start at first time
	pusher.Start()

	assert.True(t, pusher.IsRunning())

	// wait for one second in order to check whether new goroutine was generated
	time.Sleep(time.Second)

	// call Stop()
	pusher.Stop()
}

func TestPushGatewayPusher_push(t *testing.T) {
	pusher, err := NewPushGatewayPusher(
		WithIntervalMSPusher(intervalMS),
		WithRemoteAddressPusher(remoteAddr),
		WithJobNamePusher(jobName),
		WithBasicAuthPusher(basicAuth),
		WithZapLoggerEntryPusher(zapLoggerEntry),
		WithEventLoggerEntryPusher(eventLoggerEntry))
	// mock a pusher for tracing
	///doer := &HTTPDoerMock{}
	doer := atomic.Value{}
	doer.Store(&HTTPDoerMock{called: atomic.NewBool(false)})
	pusher.Pusher = pusher.Pusher.Client(doer.Load().(*HTTPDoerMock))

	assert.NotNil(t, pusher, "pusher should not be nil")
	assert.Nil(t, err, "error should be nil")

	// make state of pusher as running first
	pusher.Running.CAS(false, true)
	// run with extra go routine since push() method was an infinite loop
	go pusher.push()
	// sleep one seconds to make sure http request was called at least once
	time.Sleep(1 * time.Second)
	assert.True(t, doer.Load().(*HTTPDoerMock).called.Load(), "supposed to be called at least once")
}

func TestPushGatewayPusher_IsRunning_ExpectFalse(t *testing.T) {
	pusher, err := NewPushGatewayPusher(
		WithIntervalMSPusher(intervalMS),
		WithRemoteAddressPusher(remoteAddr),
		WithJobNamePusher(jobName),
		WithBasicAuthPusher(basicAuth),
		WithZapLoggerEntryPusher(zapLoggerEntry),
		WithEventLoggerEntryPusher(eventLoggerEntry))

	assert.NotNil(t, pusher, "pusher should not be nil")
	assert.Nil(t, err, "error should be nil")

	assert.False(t, pusher.IsRunning())

	// start periodic job
	pusher.Start()
	assert.True(t, pusher.IsRunning())
	defer pusher.Stop()
}

func TestPushGatewayPusher_IsRunning_ExpectTrue(t *testing.T) {
	pusher, err := NewPushGatewayPusher(
		WithIntervalMSPusher(intervalMS),
		WithRemoteAddressPusher(remoteAddr),
		WithJobNamePusher(jobName),
		WithBasicAuthPusher(basicAuth),
		WithZapLoggerEntryPusher(zapLoggerEntry),
		WithEventLoggerEntryPusher(eventLoggerEntry))

	assert.NotNil(t, pusher, "pusher should not be nil")
	assert.Nil(t, err, "error should be nil")

	assert.False(t, pusher.IsRunning())

	// start periodic job
	pusher.Start()
	assert.True(t, pusher.IsRunning())
	defer pusher.Stop()
}

func TestPushGatewayPusher_Stop(t *testing.T) {
	pusher, err := NewPushGatewayPusher(
		WithIntervalMSPusher(intervalMS),
		WithRemoteAddressPusher(remoteAddr),
		WithJobNamePusher(jobName),
		WithBasicAuthPusher(basicAuth),
		WithZapLoggerEntryPusher(zapLoggerEntry),
		WithEventLoggerEntryPusher(eventLoggerEntry))

	assert.NotNil(t, pusher, "pusher should not be nil")
	assert.Nil(t, err, "error should be nil")

	// start periodic job
	pusher.Start()
	assert.True(t, pusher.IsRunning())
	pusher.Stop()
	assert.False(t, pusher.IsRunning())
}

func TestPushGatewayPusher_GetPusher(t *testing.T) {
	pusher, err := NewPushGatewayPusher(
		WithIntervalMSPusher(intervalMS),
		WithRemoteAddressPusher(remoteAddr),
		WithJobNamePusher(jobName),
		WithBasicAuthPusher(basicAuth),
		WithZapLoggerEntryPusher(zapLoggerEntry),
		WithEventLoggerEntryPusher(eventLoggerEntry))

	assert.NotNil(t, pusher, "pusher should not be nil")
	assert.Nil(t, err, "error should be nil")
	assert.NotNil(t, pusher.Pusher)
}
