package objectstore

import (
	"errors"
	"testing"
)

func TestKernelSentinelsAreDistinct(t *testing.T) {
	all := []error{ErrObjectMissing, ErrHashMismatch, ErrObjectTooLarge, ErrKeyOutsideTenant}
	for i := range all {
		for j := range all {
			if i != j && errors.Is(all[i], all[j]) {
				t.Fatalf("sentinel %d and %d are not distinct", i, j)
			}
		}
	}
}
