// Copyright (c) 2008 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
// Copyright 2018 The go-matrix Authors
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

// +build !windows

package dashboard

import (
	"syscall"

	"github.com/matrix/go-matrix/log"
)

// getProcessCPUTime retrieves the process' CPU time since program startup.
func getProcessCPUTime() float64 {
	var usage syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &usage); err != nil {
		log.Warn("Failed to retrieve CPU time", "err", err)
		return 0
	}
	return float64(usage.Utime.Sec+usage.Stime.Sec) + float64(usage.Utime.Usec+usage.Stime.Usec)/1000000
}
