package isoperiod

import (
	"errors"
	"regexp"
	"strconv"
	"time"
)

var (
	compiler = regexp.MustCompile(`(R)?(\d+)?/?P(\d+Y)?(\d+M)?(\d+D)?(T)?(\d+H)?(\d+M)?(\d+S)?`)
)

// A Period represents an ISO 8601 period.
type Period struct {
	Repetitions int `json:"repetitions"`
	Year        int `json:"year"`
	Month       int `json:"month"`
	Day         int `json:"day"`
	Hour        int `json:"hour"`
	Minute      int `json:"minute"`
	Second      int `json:"second"`
	time        time.Duration
	done        chan bool
	running     bool
}

// New generates a new ISO Period.
func New(now time.Time, repitions int, year int, month int, day int, hour int, minute int, second int) *Period {
	h := time.Duration(hour) * time.Hour
	m := time.Duration(minute) * time.Minute
	s := time.Duration(second) * time.Second
	t := h + m + s

	period := &Period{
		Repetitions: repitions,
		Year:        year,
		Month:       month,
		Day:         day,
		Hour:        hour,
		Minute:      minute,
		Second:      second,
		time:        t,
	}

	return period
}

// Parse converts an ISO 8601 string to a period.
//
// Currently the following formats are supported:
// - [Rn/]P[nY][nM][nD]T[nH][nM][nS]
// - P[nY][nM][nD]T[nH][nM][nS]
//
// Examples would be:
// - P1M (1 Month, no repetitions)
// - PT1M (1 Minute, no repetitions)
// - R/PT1M (1 Minute, endless repetitions)
// - R5/PT30S (30 Seconds, 5 Times)
func Parse(s string) (*Period, error) {
	var (
		result = &Period{
			Repetitions: 0,
			time:        time.Duration(0),
		}
		err error
	)

	matches := compiler.FindStringSubmatch(s)
	if len(matches) == 0 {
		return nil, errors.New("invalid repeat format")
	}

	if matches[1] == "R" {
		result.Repetitions = -1

		if matches[2] != "" {
			result.Repetitions, err = strconv.Atoi(matches[2])
			if err != nil {
				return nil, err
			}
		}
	}

	if matches[3] != "" {
		result.Year, err = strconv.Atoi(matches[3][:len(matches[3])-1])
		if err != nil {
			return nil, err
		}
	}

	if matches[4] != "" {
		result.Month, err = strconv.Atoi(matches[4][:len(matches[4])-1])
		if err != nil {
			return nil, err
		}
	}

	if matches[5] != "" {
		result.Day, err = strconv.Atoi(matches[5][:len(matches[5])-1])
		if err != nil {
			return nil, err
		}
	}

	if matches[7] != "" {
		result.Hour, err = strconv.Atoi(matches[7][:len(matches[7])-1])
		if err != nil {
			return nil, err
		}
		result.time += time.Duration(result.Hour) * time.Hour
	}

	if matches[8] != "" {
		result.Minute, err = strconv.Atoi(matches[8][:len(matches[8])-1])
		if err != nil {
			return nil, err
		}
		result.time += time.Duration(result.Minute) * time.Minute
	}

	if matches[9] != "" {
		result.Second, err = strconv.Atoi(matches[9][:len(matches[9])-1])
		if err != nil {
			return nil, err
		}
		result.time += time.Duration(result.Second) * time.Second
	}

	return result, nil
}

// Next returns the next time the period would be valid.
// Are there no repetitions left, the result will be an empty time.Time
//
// Lowering of the "Repetitions" value is up to the user.
func (r *Period) Next(now time.Time) time.Time {
	if r.Repetitions == 0 {
		return time.Time{}
	}

	h := time.Hour * time.Duration(r.Hour)
	m := time.Minute * time.Duration(r.Minute)
	s := time.Second * time.Duration(r.Second)

	return now.AddDate(r.Year, r.Month, r.Day).Add(h + m + s)
}

// Start returns a read-only channel that triggers whenever the period becomes valid.
// After reaching the required amount of repetitions, the channel will be closed.
//
// It can also be stopped using the Stop() method.
func (r *Period) Start() <-chan time.Time {
	var c <-chan time.Time

	sender := make(chan time.Time)
	r.done = make(chan bool)
	ticker := time.NewTicker(1 * time.Second)
	timeout := time.NewTimer(5 * time.Second)
	amount := r.Repetitions

	c = sender

	go func() {
		r.running = true

		defer timeout.Stop()
		defer ticker.Stop()
		defer close(sender)
		defer close(r.done)
		defer func() {
			r.running = false
		}()

		for {
			select {
			case t := <-ticker.C:
				select {
				case sender <- t:
				default:
				}

				if amount > 0 {
					amount--
				}
				if amount == 0 {
					select {
					case r.done <- true:
					default:
					}

					return
				}
			case <-r.done:
				// End it all
				return
			}
		}
	}()

	return c
}

// Stop stops the running ticker started with the Start() method.
// If no ticker is active, it won't do anything.
func (r *Period) Stop() {
	if !r.running {
		return
	}

	r.done <- true
}

func (r *Period) String() string {
	result := ""
	timeAdded := false
	tAdded := false

	if r.Repetitions == 0 {
		result += "R/"
	} else if r.Repetitions > 0 {
		result += "R" + strconv.Itoa(r.Repetitions) + "/"
	}

	result += "P"

	if r.Year > 0 {
		result += strconv.Itoa(r.Year) + "Y"
		timeAdded = true
	}
	if r.Month > 0 {
		result += strconv.Itoa(r.Month) + "M"
		timeAdded = true
	}
	if r.Day > 0 {
		result += strconv.Itoa(r.Day) + "D"
		timeAdded = true
	}

	if r.Hour > 0 {
		result += "T" + strconv.Itoa(r.Hour) + "H"
		timeAdded = true
		tAdded = true
	}
	if r.Minute > 0 {
		if !tAdded {
			result += "T"
			tAdded = true
		}
		result += strconv.Itoa(r.Minute) + "M"
		timeAdded = true
	}
	if r.Second > 0 {
		if !tAdded {
			result += "T"
			tAdded = true
		}
		result += strconv.Itoa(r.Second) + "S"
		timeAdded = true
	}

	if !timeAdded {
		result += "T0S"
	}

	return result
}
