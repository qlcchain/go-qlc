/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package engine

import (
	"errors"
	"fmt"

	"github.com/qlcchain/go-qlc/common/types"
)

func (e *ExecuteEngine) GenerateBlock(block *types.StateBlock, signFn SignFunc) (result *GenResult, err error) {
	var sendBlock *types.StateBlock = nil
	defer func() {
		if err := recover(); err != nil {
			errDetail := fmt.Sprintf("block(addr:%v prevHash:%v )", block.Address, block.Previous)
			if sendBlock != nil {
				errDetail += fmt.Sprintf("sendBlock(addr:%v hash:%v)", block.Address, block.GetHash())
			}

			e.logger.Error(fmt.Sprintf("generator_vm panic error %v", err), "detail", errDetail)

			result = &GenResult{}
			err = errors.New("generator_vm panic error")
		}
	}()

	if block.IsReceiveBlock() {
		if sendBlock, err = e.ledger.GetStateBlock(block.Link); err != nil {
			return nil, errors.New("can not fetch block link")
		}
	}
	blockList, isRetry, err := e.Run(block, sendBlock)
	if len(blockList) > 0 {
		for k, v := range blockList {
			e.logger.Debug(k, v)
			h := v.GetHash()
			if k == 0 {
				if signFn != nil {
					signature, err := signFn(block.Address, h[:])
					if err != nil {
						return nil, err
					}
					v.Signature = signature
				}
			} else {
				blockList[k-1].Previous = h
			}
		}
	}

	return &GenResult{
		Blocks:  blockList,
		IsRetry: isRetry,
		Error:   err,
	}, nil
}

func (e *ExecuteEngine) Run(block, sendBlock *types.StateBlock) (blockList []*types.StateBlock, isRetry bool, err error) {
	return nil, false, nil
}
