// Copyright 2019 Martin Holst Swende
// This file is part of the goevmlab library.
//
// The library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the goevmlab library. If not, see <http://www.gnu.org/licenses/>.

package traces

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
)

// stack is an object for basic stack operations.
type stack struct {
	data []*common.Address
}

func newStack() *stack {
	return &stack{data: make([]*common.Address, 0, 5)}
}

func (st *stack) push(a *common.Address) {
	st.data = append(st.data, a)
}
func (st *stack) pop() (ret *common.Address) {
	if len(st.data) == 0 {
		return nil
	}
	ret = st.data[len(st.data)-1]
	st.data = st.data[:len(st.data)-1]
	return
}

// AnalyzeCalls scans through the ops, and assigns context-addresses to the
// lines.
func AnalyzeCalls(trace *Traces) {
	callStack := newStack()
	var currentAddress *common.Address
	var prevLine *TraceLine
	for _, line := range trace.Ops {
		if prevLine != nil {
			curDepth, prevDepth := line.Depth(), prevLine.Depth()
			if curDepth > prevDepth {
				// A call or create was made here
				newAddress := determineDestination(prevLine.log, currentAddress)
				callStack.push(currentAddress)
				currentAddress = newAddress
			} else if curDepth < prevDepth {
				// We backed out
				currentAddress = callStack.pop()
			}
			line.address = currentAddress
		}
		prevLine = line
	}
}

// determineDestination looks at the stack args and determines what the call
// address is
func determineDestination(log *vm.StructLog, current *common.Address) *common.Address {

	switch log.Op {
	case vm.CALL, vm.STATICCALL:
		if len(log.Stack) > 1 {
			a := common.BigToAddress(log.Stack[1])
			return &a
		}
	case vm.DELEGATECALL, vm.CALLCODE:
		// The stack index is 1, but the actual execution context remains the same
		return current
	case vm.CREATE, vm.CREATE2:
		// In order to figure this out, we would need both nonce and current address
		// while we _may_ have the address, we don't have the nonce
		return nil
	}
	return nil
}
