// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package main

import (
	"github.com/rookie-ninja/rk-prom"
	"github.com/rookie-ninja/rk-query"
	"time"
)

func main() {
	// create event data
	fac := rk_query.NewEventFactory()

	// create gin entry
	entry := rk_prom.NewPromEntry(
		rk_prom.WithPort(1608),
		rk_prom.WithPath("metrics"))

	// start server
	entry.Bootstrap(fac.CreateEvent())
	entry.Wait(1 * time.Second)
}
