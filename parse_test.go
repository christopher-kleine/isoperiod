package isoperiod_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/christopher-kleine/isoperiod"
)

var (
	now, _ = time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
)

func checkError(should, is error) error {
	if should == nil && is != nil {
		return fmt.Errorf("error is %q but should be nil", is.Error())
	}

	if should != nil && is == nil {
		return fmt.Errorf("error is nil but should be %q", should.Error())
	}

	if should == nil && is == nil {
		return nil
	}

	if should.Error() != is.Error() {
		return fmt.Errorf("error is %q but should be %q", is.Error(), should.Error())
	}

	return nil
}

func calcTime(year, month, day, hour, minute, second int) time.Time {
	h := time.Duration(hour) * time.Hour
	m := time.Duration(minute) * time.Minute
	s := time.Duration(second) * time.Second

	return now.AddDate(year, month, day).Add(h + m + s)
}

func TestParse(t *testing.T) {
	testTable := []struct {
		Repetitions int
		Year        int
		Month       int
		Day         int
		Hour        int
		Minute      int
		Second      int
		S           string
		Err         error
	}{
		{
			Repetitions: 0,
			Year:        3,
			Month:       9,
			Day:         7,
			Hour:        12,
			Minute:      30,
			Second:      50,
			S:           "P3Y9M7DT12H30M50S",
			Err:         nil,
		},
		{
			Repetitions: 0,
			Year:        0,
			Month:       1,
			Day:         12,
			Hour:        2,
			Minute:      3,
			Second:      5,
			S:           "P1M12DT2H3M5S",
			Err:         nil,
		},
		{
			Repetitions: 20,
			Year:        1,
			Month:       6,
			Day:         2,
			Hour:        2,
			Minute:      0,
			Second:      0,
			S:           "R20/P1Y6M2DT2H",
			Err:         nil,
		},
	}

	for _, testCase := range testTable {
		p, err := isoperiod.Parse(testCase.S)

		if err := checkError(testCase.Err, err); err != nil {
			t.Error(err)
			continue
		}

		if p.Repetitions != testCase.Repetitions {
			t.Errorf("Repetitions %d != %d", p.Repetitions, testCase.Repetitions)
		}

		if p.Year != testCase.Year {
			t.Errorf("Year %d != %d", p.Year, testCase.Year)
		}

		if p.Month != testCase.Month {
			t.Errorf("Month %d != %d", p.Month, testCase.Month)
		}

		if p.Day != testCase.Day {
			t.Errorf("Day %d != %d", p.Day, testCase.Day)
		}

		if p.Hour != testCase.Hour {
			t.Errorf("Hour %d != %d", p.Hour, testCase.Hour)
		}

		if p.Minute != testCase.Minute {
			t.Errorf("Minute %d != %d", p.Minute, testCase.Minute)
		}

		if p.Second != testCase.Second {
			t.Errorf("Second %d != %d", p.Second, testCase.Second)
		}
	}
}

func TestNext(t *testing.T) {
	testTable := []struct {
		Period *isoperiod.Period
		Next   time.Time
	}{
		{
			Period: isoperiod.New(now, 0, 3, 9, 7, 12, 30, 50),
			Next:   time.Time{},
		},
		{
			Period: isoperiod.New(now, 0, 0, 1, 12, 2, 3, 5),
			Next:   time.Time{},
		},
		{
			Period: isoperiod.New(now, 20, 1, 6, 2, 2, 0, 0),
			Next:   calcTime(1, 6, 2, 2, 0, 0),
		},
	}

	for _, testCase := range testTable {
		next := testCase.Period.Next(now)
		if !testCase.Next.Equal(next) {
			t.Errorf("next is %s but should be %s", next, testCase.Next)
		}
	}
}
