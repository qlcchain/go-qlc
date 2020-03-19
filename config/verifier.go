/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package config

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"gopkg.in/validator.v2"

	"github.com/qlcchain/go-qlc/common/types"
)

// qlc address validator
func addressString(v interface{}, param string) error {
	st := reflect.ValueOf(v)
	if st.Kind() != reflect.String {
		return errors.New("address only validates string")
	}
	s := st.String()
	if len(s) > 0 && !types.IsValidHexAddress(s) {
		return fmt.Errorf("invalid qlc address %s", s)
	}
	return nil
}

func hash(v interface{}, param string) error {
	st := reflect.ValueOf(v)
	h := st.Interface()

	switch v := h.(type) {
	case types.Hash:
		if v.IsZero() {
			return errors.New("hash can not be zero")
		}
	default:
		return errors.New("hash only validates types.Hash")
	}
	return nil
}

func address(v interface{}, param string) error {
	st := reflect.ValueOf(v)
	h := st.Interface()

	switch v := h.(type) {
	case types.Address:
		if v.IsZero() {
			return errors.New("address can not be zero")
		}
	default:
		return errors.New("hash only validates types.Address")
	}
	return nil
}

func init() {
	_ = validator.SetValidationFunc("address", addressString)
	_ = validator.SetValidationFunc("qlcaddress", address)
	_ = validator.SetValidationFunc("hash", hash)
}

func ErrToString(err error) string {
	if err == nil {
		return ""
	}
	if errs, ok := err.(validator.ErrorMap); ok {
		var errOuts []string
		// Iterate through the list of fields and respective errors
		errOuts = append(errOuts, "Invalid due to fields:")

		for f, e := range errs {
			errOuts = append(errOuts, fmt.Sprintf("\t - %s (%v)", f, e))
		}

		return strings.Join(errOuts, "\n")
	}

	return err.Error()
}
