package main

import (
	"fmt"
	"testing"
	"time"
)

func TestSolarHourAtIndependentOfTimezone(t *testing.T) {
	madrid, err := time.LoadLocation("Europe/Madrid")
	if err != nil {
		t.Fatal(err)
	}
	for _, date := range []string{
		"2026-01-16T22:45:00Z",
		"2026-07-16T22:45:00Z",
		"2026-10-16T22:45:00Z",
		"2026-03-29T00:59:00Z",
		"2026-03-29T01:01:00Z",
		"2026-10-25T00:59:00Z",
		"2026-10-25T01:01:00Z",
	} {
		t.Run(date, func(t *testing.T) {
			now, err := time.Parse(time.RFC3339, date)
			if err != nil {
				t.Fatal(err)
			}
			wantHour, wantMinute := solarHourAt(now, 41.38879, 2.15879)
			for _, location := range []*time.Location{
				madrid,
				time.FixedZone("UTC-8", -8*3600),
				time.FixedZone("UTC+5:30", 19800),
			} {
				hour, minute := solarHourAt(now.In(location), 41.38879, 2.15879)
				if hour != wantHour || minute != wantMinute {
					t.Errorf("%s: got %d:%02d, want %d:%02d", location, hour, minute, wantHour, wantMinute)
				}
			}
		})
	}
}

func TestSolarHourAtSolarNoon(t *testing.T) {
	date := time.Date(2026, time.October, 16, 12, 0, 0, 0, time.UTC)
	for _, longitude := range []float64{-180, -120, 0, 2.15879, 120, 180} {
		t.Run(fmt.Sprint(longitude), func(t *testing.T) {
			noon := getSolarNoon(date, 41.38879, longitude)
			for _, offset := range []time.Duration{-13*time.Hour - time.Minute, -time.Minute, 0, time.Minute, 13*time.Hour + time.Minute} {
				now := noon.Add(offset)
				// Keep the expected offset tied to this transit; the library's
				// next transit can differ by seconds, so allow one minute.
				want := time.Date(2026, time.October, 16, 12, 0, 0, 0, time.UTC).Add(offset)
				hour, minute := solarHourAt(now, 41.38879, longitude)
				difference := (hour*60 + minute - want.Hour()*60 - want.Minute() + 1440) % 1440
				if difference > 1 && difference < 1439 {
					t.Errorf("offset %s: got %d:%02d, want about %s", offset, hour, minute, want.Format("15:04"))
				}
			}
		})
	}
}

func TestSolarHourAtReportedRatHour(t *testing.T) {
	now := time.Date(2026, time.October, 17, 0, 53, 0, 0, time.FixedZone("CEST", 2*3600))
	hour, minute := solarHourAt(now, 41.38879, 2.15879)
	if hour != 23 || minute < 14 || minute > 16 {
		t.Fatalf("got %d:%02d, want approximately 23:15", hour, minute)
	}
	if idx := getAnimalIdx(hour); idx != 0 {
		t.Errorf("got animal %d, want rat (0)", idx)
	}
}

func TestGetAnimalIdx(t *testing.T) {
	for hour := 0; hour < 24; hour++ {
		want := ((hour + 1) / 2) % 12
		if got := getAnimalIdx(hour); got != want {
			t.Errorf("hour %d: got %d, want %d", hour, got, want)
		}
	}
}
