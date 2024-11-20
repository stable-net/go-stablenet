package test

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/pkg/errors"
)

type allErrorsType map[[4]byte]abi.Error

var (
	allErrorsLock = sync.RWMutex{}
	allErrors     = allErrorsType{} // eventID => abi.Event
)

func collectErrors(abi *abi.ABI) error {
	for _, e := range abi.Errors {
		func() {
			allErrorsLock.Lock()
			defer allErrorsLock.Unlock()
			sig := [4]byte{}
			copy(sig[:], e.ID[:4])
			if _, ok := allErrors[sig]; !ok {
				allErrors[sig] = e
			}
		}()
	}

	return nil
}

type RevertError struct {
	ABI    abi.Error
	Output interface{}
}

func (r *RevertError) Error() string {
	return fmt.Sprintf("%s: %s %v", vm.ErrExecutionReverted, r.ABI.Sig, r.Output)
}

// ErrorCode returns the JSON error code for a revert.
// See: https://github.com/ethereum/wiki/wiki/JSON-RPC-Error-Codes-Improvement-Proposal
func NewRevertError(err error) error {
	if revert, ok := err.(interface {
		ErrorCode() int
		ErrorData() interface{}
	}); !ok || revert.ErrorCode() != 3 {
		return err
	} else {
		if data, ok := revert.ErrorData().(string); !ok {
			return err
		} else {
			datas := hexutil.MustDecode(data)
			if revertErr, ok := UnpackError(datas); ok {
				return revertErr
			} else {
				reason, errUnpack := abi.UnpackRevert(datas)
				if errUnpack == nil {
					return fmt.Errorf("execution reverted: %v", reason)
				} else {
					return errors.New("execution reverted")
				}
			}
		}
	}
}

func UnpackError(result []byte) (error, bool) {
	sig := [4]byte{}
	copy(sig[:], result[:4])
	if errABI, ok := allErrors[sig]; !ok {
		return nil, false
	} else if output, err := errABI.Unpack(result); err != nil {
		return nil, false
	} else {
		return &RevertError{errABI, output}, true
	}
}
