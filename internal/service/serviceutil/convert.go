package serviceutil

import (
	"time"

	"github.com/shopspring/decimal"
)

func DecimalFromFloat(v float64) decimal.Decimal {
	return decimal.NewFromFloat(v)
}

func DecimalToFloat(v decimal.Decimal) float64 {
	result, _ := v.Float64()
	return result
}

func DecimalPtrToInt32Value(v *int32) int32 {
	if v == nil {
		return 0
	}
	return *v
}

func Int32PtrValue(v *int32) int32 {
	if v == nil {
		return 0
	}
	return *v
}

func Int32Ptr(v int32) *int32 {
	return &v
}

func NilIfNonPositiveInt32(v int32) *int32 {
	if v <= 0 {
		return nil
	}
	return &v
}

func TimePtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

func TimeValue(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}
