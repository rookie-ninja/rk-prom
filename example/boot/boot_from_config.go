// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package main

import (
	"github.com/rookie-ninja/rk-logger"
	"github.com/rookie-ninja/rk-prom"
	"github.com/rookie-ninja/rk-query"
	"time"
)

func main() {
	fac := rk_query.NewEventFactory()
	entry := rk_prom.NewPromEntryWithConfig("example/boot/boot.yaml", fac, rk_logger.StdoutLogger)
	entry.Bootstrap(fac.CreateEvent())
	entry.Wait(1 * time.Second)
}
