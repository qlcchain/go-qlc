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

	"github.com/qlcchain/go-qlc/common/types"
	"gopkg.in/validator.v2"
)

// qlc address validator
func address(v interface{}, param string) error {
	st := reflect.ValueOf(v)
	if st.Kind() != reflect.String {
		return errors.New("QLCAddress only validates string")
	}
	s := st.String()
	if len(s) > 0 && !types.IsValidHexAddress(s) {
		return fmt.Errorf("invalid qlc address %s", s)
	}
	return nil
}

func init() {
	_ = validator.SetValidationFunc("address", address)
}

func ErrToString(err error) string {
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
