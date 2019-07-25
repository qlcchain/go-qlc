/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package contract

import (
	"testing"
	"time"
)

func Test_timeSpan_Calculate(t *testing.T) {
	type fields struct {
		years   int
		months  int
		days    int
		hours   int
		minutes int
		seconds int
	}
	type args struct {
		t time.Time
	}
	now := time.Now()
	tests := []struct {
		name   string
		fields fields
		args   args
		want   time.Time
	}{
		{"now", fields{}, args{now}, now},
		{"6months", fields{months: 6}, args{time.Unix(1556247209, 0)}, time.Unix(1572058409, 0)},
		{"10days", fields{days: 10}, args{time.Unix(1556247263, 0)}, time.Unix(1557111263, 0)},
		{"10minutes", fields{minutes: 10}, args{time.Unix(1556247308, 0)}, time.Unix(1556247908, 0)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &timeSpan{
				years:   tt.fields.years,
				months:  tt.fields.months,
				days:    tt.fields.days,
				hours:   tt.fields.hours,
				minutes: tt.fields.minutes,
				seconds: tt.fields.seconds,
			}
			if got := ts.Calculate(tt.args.t); got.Unix() != tt.want.Unix() {
				t.Errorf("timeSpan.Calculate() input=%v, got=%v, want %v", tt.args.t, got, tt.want)
			}
		})
	}
}

func TestTimeSpan_Calculate_WithLocation(t *testing.T) {
	ms := int64(1556247209)
	t1 := time.Unix(ms, 0)
	t.Log(t1.String(), t1.UTC().String())
	ts := timeSpan{days: 180}
	t2 := ts.Calculate(t1)
	unix1 := t2.Unix()
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(loc)
	}

	t3 := t1.In(loc)
	t4 := ts.Calculate(t3)
	unix2 := t4.Unix()
	if unix1 != unix2 {
		t.Fatal("invalid unix timestamp", unix1, unix2)
	} else {
		t.Log("t1", t1.String(), t1.UTC().String())
		t.Log("t2", t2.String(), t2.UTC().String())
		t.Log("t3", t3.String(), t3.UTC().String())
		t.Log("t4", t4.String(), t4.UTC().String())
	}
}
