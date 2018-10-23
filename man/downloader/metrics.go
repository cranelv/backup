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

// Contains the metrics collected by the downloader.

package downloader

import (
	"github.com/matrix/go-matrix/metrics"
)

var (
	headerInMeter      = metrics.NewRegisteredMeter("man/downloader/headers/in", nil)
	headerReqTimer     = metrics.NewRegisteredTimer("man/downloader/headers/req", nil)
	headerDropMeter    = metrics.NewRegisteredMeter("man/downloader/headers/drop", nil)
	headerTimeoutMeter = metrics.NewRegisteredMeter("man/downloader/headers/timeout", nil)

	bodyInMeter      = metrics.NewRegisteredMeter("man/downloader/bodies/in", nil)
	bodyReqTimer     = metrics.NewRegisteredTimer("man/downloader/bodies/req", nil)
	bodyDropMeter    = metrics.NewRegisteredMeter("man/downloader/bodies/drop", nil)
	bodyTimeoutMeter = metrics.NewRegisteredMeter("man/downloader/bodies/timeout", nil)

	receiptInMeter      = metrics.NewRegisteredMeter("man/downloader/receipts/in", nil)
	receiptReqTimer     = metrics.NewRegisteredTimer("man/downloader/receipts/req", nil)
	receiptDropMeter    = metrics.NewRegisteredMeter("man/downloader/receipts/drop", nil)
	receiptTimeoutMeter = metrics.NewRegisteredMeter("man/downloader/receipts/timeout", nil)

	stateInMeter   = metrics.NewRegisteredMeter("man/downloader/states/in", nil)
	stateDropMeter = metrics.NewRegisteredMeter("man/downloader/states/drop", nil)
)
