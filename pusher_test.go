// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

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
	intervalMs       = 200 * time.Millisecond
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

func TestNewPushGatewayPusher_WithNegativeIntervalMS(t *testing.T) {
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

func TestNewPushGatewayPusher_WithEmptyURL(t *testing.T) {
	pusher, err := NewPushGatewayPusher(
		WithIntervalMSPusher(intervalMs),
		WithRemoteAddressPusher(""),
		WithJobNamePusher(jobName),
		WithBasicAuthPusher(basicAuth),
		WithZapLoggerEntryPusher(zapLoggerEntry),
		WithEventLoggerEntryPusher(eventLoggerEntry))

	assert.Nil(t, pusher, "pusher should be nil")
	assert.NotNil(t, err, "error should not be nil")
}

func TestNewPushGatewayPusher_WithEmptyJobName(t *testing.T) {
	pusher, err := NewPushGatewayPusher(
		WithIntervalMSPusher(intervalMs),
		WithRemoteAddressPusher(remoteAddr),
		WithJobNamePusher(""),
		WithBasicAuthPusher(basicAuth),
		WithZapLoggerEntryPusher(zapLoggerEntry),
		WithEventLoggerEntryPusher(eventLoggerEntry))

	assert.Nil(t, pusher, "pusher should be nil")
	assert.NotNil(t, err, "error should not be nil")
}

func TestNewPushGatewayPusher_WithNilLZapLoggerEntry(t *testing.T) {
	pusher, err := NewPushGatewayPusher(
		WithIntervalMSPusher(intervalMs),
		WithRemoteAddressPusher(remoteAddr),
		WithJobNamePusher(jobName),
		WithBasicAuthPusher(basicAuth),
		WithZapLoggerEntryPusher(nil),
		WithEventLoggerEntryPusher(eventLoggerEntry))

	assert.NotNil(t, pusher, "pusher should not be nil")
	assert.Nil(t, err, "error should be nil")
}

func TestNewPushGatewayPusher_WithNilLEventLoggerEntry(t *testing.T) {
	pusher, err := NewPushGatewayPusher(
		WithIntervalMSPusher(intervalMs),
		WithRemoteAddressPusher(remoteAddr),
		WithJobNamePusher(jobName),
		WithBasicAuthPusher(basicAuth),
		WithZapLoggerEntryPusher(zapLoggerEntry),
		WithEventLoggerEntryPusher(nil))

	assert.NotNil(t, pusher, "pusher should not be nil")
	assert.Nil(t, err, "error should be nil")
}

func TestNewPushGatewayPusher_WithEmptyBasicAuth(t *testing.T) {
	pusher, err := NewPushGatewayPusher(
		WithIntervalMSPusher(intervalMs),
		WithRemoteAddressPusher(remoteAddr),
		WithJobNamePusher(jobName),
		WithBasicAuthPusher(""),
		WithZapLoggerEntryPusher(zapLoggerEntry),
		WithEventLoggerEntryPusher(eventLoggerEntry))

	assert.NotNil(t, pusher, "pusher should not be nil")
	assert.Nil(t, err, "error should be nil")
}

func TestNewPushGatewayPusher_WithInvalidBasicAuth(t *testing.T) {
	pusher, err := NewPushGatewayPusher(
		WithIntervalMSPusher(intervalMs),
		WithRemoteAddressPusher(remoteAddr),
		WithJobNamePusher(jobName),
		WithBasicAuthPusher("user:pass:pass"),
		WithZapLoggerEntryPusher(zapLoggerEntry),
		WithEventLoggerEntryPusher(eventLoggerEntry))

	assert.NotNil(t, pusher, "pusher should not be nil")
	assert.Nil(t, err, "error should be nil")
}

func TestNewPushGatewayPusher_WithCert(t *testing.T) {
	serverCert := `
-----BEGIN CERTIFICATE-----
MIIC/jCCAeagAwIBAgIUWVMP53O835+njsr23UZIX2KEXGYwDQYJKoZIhvcNAQEL
BQAwYDELMAkGA1UEBhMCQ04xEDAOBgNVBAgTB0JlaWppbmcxCzAJBgNVBAcTAkJK
MQswCQYDVQQKEwJSSzEQMA4GA1UECxMHUksgRGVtbzETMBEGA1UEAxMKUksgRGVt
byBDQTAeFw0yMTA0MDcxMzAzMDBaFw0yNjA0MDYxMzAzMDBaMEIxCzAJBgNVBAYT
AkNOMRAwDgYDVQQIEwdCZWlqaW5nMQswCQYDVQQHEwJCSjEUMBIGA1UEAxMLZXhh
bXBsZS5uZXQwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAARf8p/nxvY1HHUkJXZk
fFQgDtQ2CK9DOAe6y3lE21HTJ/Vi4vHNqWko9koyYgKqgUXyiq5lGAswo68KvmD7
c2L4o4GYMIGVMA4GA1UdDwEB/wQEAwIFoDATBgNVHSUEDDAKBggrBgEFBQcDATAM
BgNVHRMBAf8EAjAAMB0GA1UdDgQWBBTv6dUlEI6NcQBzihnzKZrxKpbnTTAfBgNV
HSMEGDAWgBRgwpYKhgfeO3p2XuX0he35caeUgTAgBgNVHREEGTAXgglsb2NhbGhv
c3SHBH8AAAGHBAAAAAAwDQYJKoZIhvcNAQELBQADggEBAByqLc3QkaGNr+QqjFw7
znk9j0X4Ucm/1N6iGIp8fUi9t+mS1La6CB1ej+FoWkSYskzqBpdIkqzqZan1chyF
njhtMsWgZYW6srXNRgByA9XS2s28+xg9owcpceXa3wG4wbnTj1emcunzSrKVFjS1
IJUjl5HWCKibnVjgt4g0s9tc8KYpXkGYl23U4FUta/07YFmtW5SDF38NWrNOe5qV
EALMz1Ry0PMgY0SDtKhddDNnNS32fz40IP0wB7a31T24eZetZK/INaIi+5SM0iLx
kfqN71xKxAIIYmuI9YwWCFaZ2+qbLIiDTbR6gyuLIQ2AfwBLZ06g939ZfSqZuP8P
oxU=
-----END CERTIFICATE-----
`
	serverKey := `
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIPSAlU9WxdGhhmdJqg3OLmUPZlnKhejtZ2LbFNBkCTJfoAoGCCqGSM49
AwEHoUQDQgAEX/Kf58b2NRx1JCV2ZHxUIA7UNgivQzgHust5RNtR0yf1YuLxzalp
KPZKMmICqoFF8oquZRgLMKOvCr5g+3Ni+A==
-----END EC PRIVATE KEY-----
`

	certStore := &rkentry.CertStore{
		ServerCert: []byte(serverCert),
		ServerKey:  []byte(serverKey),
	}

	pusher, err := NewPushGatewayPusher(
		WithCertStorePusher(certStore),
		WithIntervalMSPusher(intervalMs),
		WithRemoteAddressPusher(remoteAddr),
		WithJobNamePusher(jobName),
		WithBasicAuthPusher(basicAuth),
		WithZapLoggerEntryPusher(zapLoggerEntry),
		WithEventLoggerEntryPusher(eventLoggerEntry))

	assert.NotNil(t, pusher, "pusher should not be nil")
	assert.Nil(t, err, "error should be nil")

	assert.Equal(t, intervalMs, pusher.IntervalMs)
	assert.NotNil(t, pusher.Pusher, "pusher should not be nil")
	assert.Equal(t, jobName, pusher.JobName)
	assert.Equal(t, basicAuth, pusher.Credential)
	assert.NotNil(t, pusher.lock, "lock should not be nil")
	assert.False(t, pusher.Running.Load(), "isRunning should be false")
	assert.Contains(t, pusher.RemoteAddress, "https")
}

func TestNewPushGatewayPusher_WithInvalidCert(t *testing.T) {
	serverCert := `Invalid`
	serverKey := `Invalid`

	certStore := &rkentry.CertStore{
		ServerCert: []byte(serverCert),
		ServerKey:  []byte(serverKey),
	}

	pusher, err := NewPushGatewayPusher(
		WithCertStorePusher(certStore),
		WithIntervalMSPusher(intervalMs),
		WithRemoteAddressPusher(remoteAddr),
		WithJobNamePusher(jobName),
		WithBasicAuthPusher(basicAuth),
		WithZapLoggerEntryPusher(zapLoggerEntry),
		WithEventLoggerEntryPusher(eventLoggerEntry))

	assert.NotNil(t, pusher, "pusher should not be nil")
	assert.Nil(t, err, "error should be nil")

	assert.Equal(t, intervalMs, pusher.IntervalMs)
	assert.NotNil(t, pusher.Pusher, "pusher should not be nil")
	assert.Equal(t, jobName, pusher.JobName)
	assert.Equal(t, basicAuth, pusher.Credential)
	assert.NotNil(t, pusher.lock, "lock should not be nil")
	assert.False(t, pusher.Running.Load(), "isRunning should be false")
	assert.Contains(t, pusher.RemoteAddress, "http")
}

func TestNewPushGatewayPusher_HappyCase(t *testing.T) {
	pusher, err := NewPushGatewayPusher(
		WithIntervalMSPusher(intervalMs),
		WithRemoteAddressPusher(remoteAddr),
		WithJobNamePusher(jobName),
		WithBasicAuthPusher(basicAuth),
		WithZapLoggerEntryPusher(zapLoggerEntry),
		WithEventLoggerEntryPusher(eventLoggerEntry))

	assert.NotNil(t, pusher, "pusher should not be nil")
	assert.Nil(t, err, "error should be nil")

	assert.Equal(t, intervalMs, pusher.IntervalMs)
	assert.NotNil(t, pusher.Pusher, "pusher should not be nil")
	assert.Equal(t, jobName, pusher.JobName)
	assert.Equal(t, basicAuth, pusher.Credential)
	assert.NotNil(t, pusher.lock, "lock should not be nil")
	assert.False(t, pusher.Running.Load(), "isRunning should be false")
}

func TestPushGatewayPusher_Start_WithDuplicateStartCalls(t *testing.T) {
	pusher, err := NewPushGatewayPusher(
		WithIntervalMSPusher(intervalMs),
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
		WithIntervalMSPusher(intervalMs),
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
		WithIntervalMSPusher(intervalMs),
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
		WithIntervalMSPusher(intervalMs),
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
		WithIntervalMSPusher(intervalMs),
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
		WithIntervalMSPusher(intervalMs),
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
		WithIntervalMSPusher(intervalMs),
		WithRemoteAddressPusher(remoteAddr),
		WithJobNamePusher(jobName),
		WithBasicAuthPusher(basicAuth),
		WithZapLoggerEntryPusher(zapLoggerEntry),
		WithEventLoggerEntryPusher(eventLoggerEntry))

	assert.NotNil(t, pusher, "pusher should not be nil")
	assert.Nil(t, err, "error should be nil")
	assert.NotNil(t, pusher.GetPusher())
}
