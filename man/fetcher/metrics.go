// Copyright (c) 2008 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
// Copyright 2015 The go-matrix Authors
// This file is part of the go-matrix library.
//
// The go-matrix library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-matrix library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-matrix library. If not, see <http://www.gnu.org/licenses/>.

// Contains the metrics collected by the fetcher.

package fetcher

import (
	"github.com/matrix/go-matrix/metrics"
)

var (
	propAnnounceInMeter   = metrics.NewRegisteredMeter("man/fetcher/prop/announces/in", nil)
	propAnnounceOutTimer  = metrics.NewRegisteredTimer("man/fetcher/prop/announces/out", nil)
	propAnnounceDropMeter = metrics.NewRegisteredMeter("man/fetcher/prop/announces/drop", nil)
	propAnnounceDOSMeter  = metrics.NewRegisteredMeter("man/fetcher/prop/announces/dos", nil)

	propBroadcastInMeter   = metrics.NewRegisteredMeter("man/fetcher/prop/broadcasts/in", nil)
	propBroadcastOutTimer  = metrics.NewRegisteredTimer("man/fetcher/prop/broadcasts/out", nil)
	propBroadcastDropMeter = metrics.NewRegisteredMeter("man/fetcher/prop/broadcasts/drop", nil)
	propBroadcastDOSMeter  = metrics.NewRegisteredMeter("man/fetcher/prop/broadcasts/dos", nil)

	headerFetchMeter = metrics.NewRegisteredMeter("man/fetcher/fetch/headers", nil)
	bodyFetchMeter   = metrics.NewRegisteredMeter("man/fetcher/fetch/bodies", nil)

	headerFilterInMeter  = metrics.NewRegisteredMeter("man/fetcher/filter/headers/in", nil)
	headerFilterOutMeter = metrics.NewRegisteredMeter("man/fetcher/filter/headers/out", nil)
	bodyFilterInMeter    = metrics.NewRegisteredMeter("man/fetcher/filter/bodies/in", nil)
	bodyFilterOutMeter   = metrics.NewRegisteredMeter("man/fetcher/filter/bodies/out", nil)
)
