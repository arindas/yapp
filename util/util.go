package util

import (
	testing "testing"
)

func ActionNecessary(err error, t *testing.T) bool {
	if err != nil {
		t.Errorf("%v", err)
		return true
	}
	return false
}
