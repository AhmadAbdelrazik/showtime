package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSchedule_IsFree(t *testing.T) {
	schedule := Schedule{
		From: time.Date(2025, time.December, 5, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2025, time.December, 12, 0, 0, 0, 0, time.UTC),
	}

	shows := []Show{
		{
			StartTime: time.Date(2025, time.December, 4, 22, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, time.December, 5, 1, 0, 0, 0, time.UTC),
		},
		{
			StartTime: time.Date(2025, time.December, 6, 12, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, time.December, 6, 15, 0, 0, 0, time.UTC),
		},
		{
			StartTime: time.Date(2025, time.December, 6, 23, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, time.December, 7, 2, 0, 0, 0, time.UTC),
		},
		{
			StartTime: time.Date(2025, time.December, 7, 15, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, time.December, 7, 18, 0, 0, 0, time.UTC),
		},
		{
			StartTime: time.Date(2025, time.December, 11, 22, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, time.December, 12, 3, 0, 0, 0, time.UTC),
		},
	}

	schedule.Shows = shows

	tests := []struct {
		name               string
		startTime, endTime time.Time
		want               error
	}{
		{
			name:      "valid show",
			startTime: time.Date(2025, time.December, 5, 3, 0, 0, 0, time.UTC),
			endTime:   time.Date(2025, time.December, 5, 6, 0, 0, 0, time.UTC),
			want:      nil,
		},
		{
			name:      "before schedule",
			startTime: time.Date(2025, time.December, 4, 3, 0, 0, 0, time.UTC),
			endTime:   time.Date(2025, time.December, 4, 6, 0, 0, 0, time.UTC),
			want:      ErrInvalidSchedule,
		},
		{
			name:      "after schedule",
			startTime: time.Date(2025, time.December, 12, 3, 0, 0, 0, time.UTC),
			endTime:   time.Date(2025, time.December, 12, 6, 0, 0, 0, time.UTC),
			want:      ErrInvalidSchedule,
		},
		{
			name:      "collision",
			startTime: time.Date(2025, time.December, 6, 14, 0, 0, 0, time.UTC),
			endTime:   time.Date(2025, time.December, 6, 17, 0, 0, 0, time.UTC),
			want:      ErrInvalidSchedule,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := schedule.IsFree(Show{
				StartTime: tc.startTime,
				EndTime:   tc.endTime,
			})
			if tc.want != nil {
				assert.ErrorIs(t, err, ErrInvalidSchedule)
			} else {
				assert.Nil(t, err)
			}
		})
	}

}
