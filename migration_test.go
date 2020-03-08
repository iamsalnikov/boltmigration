package migration

import (
	"errors"
	"fmt"
	"testing"

	"go.etcd.io/bbolt"
)

func Test_NewMigrationNames(t *testing.T) {
	testError := errors.New("test error")
	type testCase struct {
		newNames      map[string]UpFunc
		appliedNames  []string
		applyErr      error
		expectedErr   error
		expectedNames []string
	}

	testCases := []testCase{
		{
			newNames:      map[string]UpFunc{},
			appliedNames:  []string{},
			applyErr:      nil,
			expectedErr:   nil,
			expectedNames: []string{},
		},
		{
			newNames: map[string]UpFunc{
				"0": func(db *bbolt.DB) error { return nil },
			},
			appliedNames:  []string{},
			applyErr:      nil,
			expectedErr:   nil,
			expectedNames: []string{"0"},
		},
		{
			newNames: map[string]UpFunc{
				"0": func(db *bbolt.DB) error { return nil },
			},
			appliedNames:  []string{"0"},
			applyErr:      nil,
			expectedErr:   nil,
			expectedNames: []string{},
		},
		{
			newNames: map[string]UpFunc{
				"0": func(db *bbolt.DB) error { return nil },
				"1": func(db *bbolt.DB) error { return nil },
				"2": func(db *bbolt.DB) error { return nil },
			},
			appliedNames:  []string{"0", "2"},
			applyErr:      nil,
			expectedErr:   nil,
			expectedNames: []string{"1"},
		},
		{
			newNames: map[string]UpFunc{
				"2": func(db *bbolt.DB) error { return nil },
				"1": func(db *bbolt.DB) error { return nil },
				"0": func(db *bbolt.DB) error { return nil },
			},
			appliedNames:  []string{"0"},
			applyErr:      nil,
			expectedErr:   nil,
			expectedNames: []string{"1", "2"},
		},
		{
			newNames: map[string]UpFunc{
				"0": func(db *bbolt.DB) error { return nil },
			},
			appliedNames:  []string{"2"},
			applyErr:      testError,
			expectedErr:   testError,
			expectedNames: []string{},
		},
	}

	for i, c := range testCases {
		resetMigrations()
		resetAppliedFunc()
		resetMarkAppliedFunc()

		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			getApplied = AppliedFunc(func(db *bbolt.DB) ([]string, error) {
				return c.appliedNames, c.applyErr
			})

			for name, mig := range c.newNames {
				Add(name, mig)
			}

			result, err := NewMigrationNames()
			if err != c.expectedErr {
				t.Errorf("I expected to get error \"%v\" but got \"%v\"", c.expectedErr, err)
				return
			}

			if !isEqualSlices(result, c.expectedNames) {
				t.Errorf("I expected to get names %v but got %v", c.expectedNames, result)
			}
		})
	}
}

func isEqualSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for j, n := range a {
		if n != b[j] {
			return false
		}
	}

	return true
}
