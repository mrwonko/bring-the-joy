package main

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"
)

func main() {
	const (
		numJoy         = 4
		minJoyDistance = time.Hour
		joyMessage     = "bring the joy"
		joySeed        = 1337
		joyUUID        = "4102840E-9CFC-4B12-89DE-F7A00133753C" // use a fixed UUID
	)
	dailyJoyWindows := []interval{
		{start: time.Date(2020, time.February, 22, 10, 0, 0, 0, time.UTC), end: time.Date(2020, time.February, 22, 22, 0, 0, 0, time.UTC)},
		{start: time.Date(2020, time.February, 23, 8, 0, 0, 0, time.UTC), end: time.Date(2020, time.February, 23, 22, 0, 0, 0, time.UTC)},
		{start: time.Date(2020, time.February, 24, 7, 0, 0, 0, time.UTC), end: time.Date(2020, time.February, 24, 22, 0, 0, 0, time.UTC)},
		{start: time.Date(2020, time.February, 25, 7, 0, 0, 0, time.UTC), end: time.Date(2020, time.February, 25, 22, 0, 0, 0, time.UTC)},
		{start: time.Date(2020, time.February, 26, 7, 0, 0, 0, time.UTC), end: time.Date(2020, time.February, 26, 22, 0, 0, 0, time.UTC)},
		{start: time.Date(2020, time.February, 27, 7, 0, 0, 0, time.UTC), end: time.Date(2020, time.February, 27, 22, 0, 0, 0, time.UTC)},
		{start: time.Date(2020, time.February, 28, 7, 0, 0, 0, time.UTC), end: time.Date(2020, time.February, 28, 22, 0, 0, 0, time.UTC)},
	}
	var joyTimes []time.Time
	rnd := rand.New(rand.NewSource(joySeed))
	for _, window := range dailyJoyWindows {
		joyTimes = append(joyTimes, spreadTheJoy(rnd, window, numJoy, minJoyDistance)...)
	}
	if err := writeJoyTimesICalendar(os.Stdout, joyTimes, joyMessage, minJoyDistance, joyUUID); err != nil {
		if _, err2 := fmt.Fprintf(os.Stderr, "failed to write joy times: %s", err); err2 != nil {
			panic(fmt.Sprintf("failed to write error %q to stderr: %s", err, err2))
		}
	}
}

func writeJoyTimesICalendar(w io.Writer, joyTimes []time.Time, joyMessage string, joyDuration time.Duration, joyUUID string) error {
	now := time.Now()
	if strings.Contains(joyMessage, "\r\n") {
		return fmt.Errorf("don't be evil")
	}
	_, err := fmt.Fprint(w, "BEGIN:VCALENDAR\r\n")
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(w, "VERSION:2.0\r\n")
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(w, "PRODID:de.mrwonko.bringthejoy@1.0\r\n")
	if err != nil {
		return err
	}
	for i, joyTime := range joyTimes {
		_, err = fmt.Fprint(w, "BEGIN:VEVENT\r\n")
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(w, "DTSTAMP:", encodeTime(now), "\r\n")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(w, "UID:%s-%d\r\n", joyUUID, i)
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(w, "TRANSP:TRANSPARENT\r\n") // don't block time
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(w, "DTSTART:", encodeTime(joyTime), "\r\n")
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(w, "DTEND:", encodeTime(joyTime.Add(joyDuration)), "\r\n")
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(w, "SUMMARY:", joyMessage, "\r\n")
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(w, "BEGIN:VALARM\r\n")
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(w, "ACTION:DISPLAY\r\n")
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(w, "DESCRIPTION:", joyMessage, "\r\n")
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(w, "TRIGGER:-PT1M\r\n")
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(w, "END:VALARM\r\n")
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(w, "END:VEVENT\r\n")
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprint(w, "END:VCALENDAR\r\n")
	if err != nil {
		return err
	}
	return nil
}

func encodeTime(t time.Time) string {
	return t.UTC().Format("20060102T150405Z")
}

type interval struct {
	// invariant: start <= end
	// end is exclusive
	start, end time.Time
}

func (i interval) GoString() string {
	return fmt.Sprintf("[%s,%s)", i.start.Format("15:04:05"), i.end.Format("15:04:05"))
}

func (i interval) contains(t time.Time) bool {
	return !t.Before(i.start) && // start is inclusive
		i.end.After(t) // end is exclusive
}

func (i interval) length() time.Duration {
	return i.end.Sub(i.start)
}

func intervalAround(center time.Time, halfWidth time.Duration) interval {
	return interval{
		start: center.Add(-halfWidth),
		end:   center.Add(halfWidth),
	}
}

func difference(minuend interval, subtrahend interval) []interval {
	// it feels like there's a simpler way to do this, but I think this should work?
	switch {
	case subtrahend.contains(minuend.start) && subtrahend.contains(minuend.end):
		return nil
	case subtrahend.contains(minuend.start):
		return []interval{{start: subtrahend.end, end: minuend.end}}
	case minuend.contains(subtrahend.start) && minuend.contains(subtrahend.end):
		return []interval{
			{start: minuend.start, end: subtrahend.start},
			{start: subtrahend.end, end: minuend.end},
		}
	case minuend.contains(subtrahend.start):
		return []interval{{start: minuend.start, end: subtrahend.start}}
	default:
		return []interval{minuend}
	}
}

func differences(minuends []interval, subtrahend interval) []interval {
	var res []interval
	for _, minuend := range minuends {
		res = append(res, difference(minuend, subtrahend)...)
	}
	return res
}

// randomPointIn should receive non-overlapping intervals or the result won't be fair.
// There must be at least one interval.
func randomPointIn(rnd *rand.Rand, intervals []interval) time.Time {
	var movingSum []time.Duration
	total := time.Duration(0)
	for _, interval := range intervals {
		total += interval.length()
		movingSum = append(movingSum, total)
	}
	offset := time.Duration(rnd.Int63n(int64(total)))
	var i int
	var sum time.Duration
	for i, sum = range movingSum {
		if offset >= sum {
			continue
		}
		break
	}
	return intervals[i].end.Add(offset - sum)
}

func spreadTheJoy(rnd *rand.Rand, within interval, numJoy int, minJoyDistance time.Duration) []time.Time {
	intervals := []interval{within}
	res := make([]time.Time, 0, numJoy)
	for i := 0; i < numJoy && len(intervals) > 0; i++ {
		joyTime := randomPointIn(rnd, intervals)
		res = append(res, joyTime)
		joyBuffer := intervalAround(joyTime, minJoyDistance)
		intervals = differences(intervals, joyBuffer)
		//fmt.Fprintf(os.Stderr, "joy @ %s, remaining windows: %#v\n", joyTime.Format("15:04:05"), intervals)
	}
	return res
}
