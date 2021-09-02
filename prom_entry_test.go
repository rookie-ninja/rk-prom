// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package rkprom

import (
	"context"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strconv"
	"testing"
	"time"
)

const bootFile = `
---
prom:
  enabled: true
  port: 1608
  path: metrics
  pusher:
    enabled: true
    intervalMS: 1000
    jobName: "rk-job"
    remoteAddress: "localhost:9091"
    basicAuth: "user:pass"
`

func TestWithPort_HappyCase(t *testing.T) {
	entry := RegisterPromEntry(
		WithPort(1949),
		WithZapLoggerEntry(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntry(rkentry.NoopEventLoggerEntry()))
	assert.Equal(t, uint64(1949), entry.Port)
}

func TestWithPath_HappyCase(t *testing.T) {
	entry := RegisterPromEntry(
		WithPath("/metrics"),
		WithZapLoggerEntry(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntry(rkentry.NoopEventLoggerEntry()))
	assert.Equal(t, "/metrics", entry.Path)
}

//func TestWithLogger_HappyCase(t *testing.T) {
//	logger := rklogger.StdoutLogger
//	entry := NewPromEntry(WithLogger(logger))
//	assert.Equal(t, logger, entry.logger)
//}
//
//func TestWithEventFactory_HappyCase(t *testing.T) {
//	fac := rkquery.NewEventFactory()
//	entry := NewPromEntry(WithEventFactory(fac))
//	assert.Equal(t, fac, entry.factory)
//}

func TestWithPusher_HappyCase(t *testing.T) {
	pusher, err := NewPushGatewayPusher(
		WithIntervalMSPusher(time.Second),
		WithRemoteAddressPusher("localhost"),
		WithJobNamePusher("job"),
		WithBasicAuthPusher("user:pass"),
		WithZapLoggerEntryPusher(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntryPusher(rkentry.NoopEventLoggerEntry()))

	assert.Nil(t, err)
	assert.NotNil(t, pusher)

	entry := RegisterPromEntry(WithPusher(pusher))
	assert.Equal(t, pusher, entry.Pusher)
}

func TestNewPromEntryWithConfig_WithEmptyString(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, true)
		} else {
			// this should never be called in case of a bug
			assert.True(t, false)
		}
	}()

	RegisterPromEntriesWithConfig("")
}

func TestNewPromEntryWithConfig_WithNonExistFile(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, true)
		} else {
			// this should never be called in case of a bug
			assert.True(t, false)
		}
	}()

	RegisterPromEntriesWithConfig("non-exist-file")
}

func TestNewPromEntryWithConfig_WithNilEventFactory(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	// create file
	configFilePath := path.Join(t.TempDir(), "boot.yaml")
	assert.Nil(t, ioutil.WriteFile(configFilePath, []byte(bootFile), os.ModePerm))
	entries := RegisterPromEntriesWithConfig(configFilePath)
	assert.True(t, len(entries) == 1)
	assert.NotNil(t, entries[PromEntryNameDefault])
	// validate prom entry with config
	entry := entries[PromEntryNameDefault].(*PromEntry)
	// ---
	// prom:
	//   enabled: true
	//   port: 1608
	//   path: metrics
	//   pusher:
	//     enabled: true
	//     intervalMS: 1000
	//     jobName: "rk-job"
	//     remoteAddress: "localhost:9091"
	//     basicAuth: "user:pass"
	assert.Equal(t, PromEntryType, entry.GetType())
	assert.Equal(t, uint64(1608), entry.Port)
	assert.Equal(t, "/metrics", entry.Path)
	assert.NotNil(t, entry.ZapLoggerEntry)
	assert.NotNil(t, entry.EventLoggerEntry)
	assert.NotNil(t, entry.Pusher)
	assert.Equal(t, time.Duration(1000)*time.Millisecond, entry.Pusher.IntervalMs)
	assert.Equal(t, "rk-job", entry.Pusher.JobName)
	assert.Equal(t, "localhost:9091", entry.Pusher.RemoteAddress)
	assert.Equal(t, "user:pass", entry.Pusher.Credential)
}

func TestNewPromEntryWithConfig_WithNilLogger(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	// create file
	configFilePath := path.Join(t.TempDir(), "boot.yaml")
	assert.Nil(t, ioutil.WriteFile(configFilePath, []byte(bootFile), os.ModePerm))
	entries := RegisterPromEntriesWithConfig(configFilePath)
	assert.True(t, len(entries) == 1)
	assert.NotNil(t, entries[PromEntryNameDefault])
	// validate prom entry with config
	entry := entries[PromEntryNameDefault].(*PromEntry)
	// ---
	// prom:
	//   enabled: true
	//   port: 1608
	//   path: metrics
	//   pusher:
	//     enabled: true
	//     intervalMS: 1000
	//     jobName: "rk-job"
	//     remoteAddress: "localhost:9091"
	//     basicAuth: "user:pass"
	assert.Equal(t, PromEntryType, entry.GetType())
	assert.Equal(t, uint64(1608), entry.Port)
	assert.Equal(t, "/metrics", entry.Path)
	assert.NotNil(t, entry.ZapLoggerEntry)
	assert.NotNil(t, entry.EventLoggerEntry)
	assert.NotNil(t, entry.Pusher)
	assert.Equal(t, time.Duration(1000)*time.Millisecond, entry.Pusher.IntervalMs)
	assert.Equal(t, "rk-job", entry.Pusher.JobName)
	assert.Equal(t, "localhost:9091", entry.Pusher.RemoteAddress)
	assert.Equal(t, "user:pass", entry.Pusher.Credential)
}

func TestNewPromEntryWithConfig_HappyCase(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	// create file
	configFilePath := path.Join(t.TempDir(), "boot.yaml")
	assert.Nil(t, ioutil.WriteFile(configFilePath, []byte(bootFile), os.ModePerm))
	entries := RegisterPromEntriesWithConfig(configFilePath)
	assert.True(t, len(entries) == 1)
	assert.NotNil(t, entries[PromEntryNameDefault])
	// validate prom entry with config
	entry := entries[PromEntryNameDefault].(*PromEntry)
	// ---
	// prom:
	//   enabled: true
	//   port: 1608
	//   path: metrics
	//   pusher:
	//     enabled: true
	//     intervalMS: 1000
	//     jobName: "rk-job"
	//     remoteAddress: "localhost:9091"
	//     basicAuth: "user:pass"
	assert.Equal(t, PromEntryType, entry.GetType())
	assert.Equal(t, uint64(1608), entry.Port)
	assert.Equal(t, "/metrics", entry.Path)
	assert.NotNil(t, entry.ZapLoggerEntry)
	assert.NotNil(t, entry.EventLoggerEntry)
	assert.NotNil(t, entry.Pusher)
	assert.Equal(t, time.Duration(1000)*time.Millisecond, entry.Pusher.IntervalMs)
	assert.Equal(t, "rk-job", entry.Pusher.JobName)
	assert.Equal(t, "localhost:9091", entry.Pusher.RemoteAddress)
	assert.Equal(t, "user:pass", entry.Pusher.Credential)
}

func TestNewPromEntry_HappyCase(t *testing.T) {
	entry := RegisterPromEntry(
		WithZapLoggerEntry(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntry(rkentry.NoopEventLoggerEntry()))
	assert.NotNil(t, entry)

	assert.Nil(t, entry.Pusher)
	assert.Equal(t, defaultPort, entry.Port)
	assert.Equal(t, defaultPath, entry.Path)
	assert.Equal(t, PromEntryNameDefault, entry.EntryName)
	assert.Equal(t, PromEntryType, entry.EntryType)
	assert.NotNil(t, rkentry.GlobalAppCtx.GetEntry(PromEntryNameDefault))
}

func TestNewPromEntry_WithPort(t *testing.T) {
	entry := RegisterPromEntry(
		WithPort(2021),
		WithZapLoggerEntry(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntry(rkentry.NoopEventLoggerEntry()))

	assert.NotNil(t, entry)

	assert.Nil(t, entry.Pusher)
	assert.Equal(t, uint64(2021), entry.Port)
	assert.Equal(t, defaultPath, entry.Path)
	assert.Equal(t, PromEntryNameDefault, entry.EntryName)
	assert.Equal(t, PromEntryType, entry.EntryType)
	assert.NotNil(t, rkentry.GlobalAppCtx.GetEntry(PromEntryNameDefault))
}

func TestNewPromEntry_WithPath(t *testing.T) {
	entry := RegisterPromEntry(
		WithPath("path"),
		WithZapLoggerEntry(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntry(rkentry.NoopEventLoggerEntry()))
	assert.NotNil(t, entry)

	assert.Nil(t, entry.Pusher)
	assert.Equal(t, defaultPort, entry.Port)
	assert.Equal(t, "/path", entry.Path)
	assert.Equal(t, PromEntryNameDefault, entry.EntryName)
	assert.Equal(t, PromEntryType, entry.EntryType)
	assert.NotNil(t, rkentry.GlobalAppCtx.GetEntry(PromEntryNameDefault))
}

func TestNewPromEntry_WithPusher(t *testing.T) {
	pusher, err := NewPushGatewayPusher(
		WithIntervalMSPusher(time.Second),
		WithRemoteAddressPusher("localhost"),
		WithJobNamePusher("job"),
		WithBasicAuthPusher("user:pass"),
		WithZapLoggerEntryPusher(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntryPusher(rkentry.NoopEventLoggerEntry()))

	assert.Nil(t, err)

	entry := RegisterPromEntry(
		WithPusher(pusher),
		WithZapLoggerEntry(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntry(rkentry.NoopEventLoggerEntry()))
	assert.NotNil(t, entry)

	assert.Equal(t, pusher, entry.Pusher)
	assert.Equal(t, defaultPort, entry.Port)
	assert.Equal(t, defaultPath, entry.Path)
	assert.Equal(t, PromEntryNameDefault, entry.EntryName)
	assert.Equal(t, PromEntryType, entry.EntryType)
	assert.NotNil(t, rkentry.GlobalAppCtx.GetEntry(PromEntryNameDefault))
}

func TestPromEntry_Bootstrap_HappyCase(t *testing.T) {
	entry := RegisterPromEntry(
		WithZapLoggerEntry(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntry(rkentry.NoopEventLoggerEntry()))
	entry.Bootstrap(context.Background())
	defer entry.Interrupt(context.Background())

	// wait for 100 milliseconds for prom client start
	time.Sleep(100 * time.Millisecond)

	assert.NotNil(t, entry.Server)
	validateServerIsUp(t, entry.Port)
}

func TestPromEntry_Shutdown_HappyCase(t *testing.T) {
	entry := RegisterPromEntry(
		WithZapLoggerEntry(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntry(rkentry.NoopEventLoggerEntry()))
	entry.Bootstrap(context.Background())

	// wait for 100 milliseconds for prom client start
	time.Sleep(100 * time.Millisecond)

	assert.NotNil(t, entry.Server)
	validateServerIsUp(t, entry.Port)

	entry.Interrupt(context.Background())
	validateServerIsDown(t, entry.Port)
}

func TestPromEntry_Shutdown_WithNilEvent(t *testing.T) {
	entry := RegisterPromEntry(
		WithZapLoggerEntry(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntry(rkentry.NoopEventLoggerEntry()))
	entry.Bootstrap(context.Background())

	// wait for 100 milliseconds for prom client start
	time.Sleep(100 * time.Millisecond)

	assert.NotNil(t, entry.Server)
	validateServerIsUp(t, entry.Port)

	entry.Interrupt(context.Background())
	validateServerIsDown(t, entry.Port)
}

func TestPromEntry_GetName_HappyCase(t *testing.T) {
	assert.Equal(t, PromEntryNameDefault, RegisterPromEntry(
		WithZapLoggerEntry(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntry(rkentry.NoopEventLoggerEntry())).GetName())
}

func TestPromEntry_GetType_HappyCase(t *testing.T) {
	assert.Equal(t, PromEntryType, RegisterPromEntry(
		WithZapLoggerEntry(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntry(rkentry.NoopEventLoggerEntry())).GetType())
}

func TestPromEntry_String_HappyCase(t *testing.T) {
	// create full prom client from config
	configFilePath := path.Join(t.TempDir(), "boot.yaml")
	assert.Nil(t, ioutil.WriteFile(configFilePath, []byte(bootFile), os.ModePerm))
	entries := RegisterPromEntriesWithConfig(configFilePath)
	assert.True(t, len(entries) == 1)
	entry := entries[PromEntryNameDefault]
	assert.NotNil(t, entry)
	assert.NotEmpty(t, entry.String())
	assert.NotEmpty(t, entry.(*PromEntry).Pusher.String())
}

func validateServerIsUp(t *testing.T, port uint64) {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("0.0.0.0", strconv.FormatUint(port, 10)), time.Second)
	assert.Nil(t, err)
	assert.NotNil(t, conn)
	if conn != nil {
		assert.Nil(t, conn.Close())
	}
}

func validateServerIsDown(t *testing.T, port uint64) {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("0.0.0.0", strconv.FormatUint(port, 10)), time.Second)
	assert.NotNil(t, err)
	assert.Nil(t, conn)
}
