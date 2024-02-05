package ad

import (
	"context"
	"testing"
	"time"

	"encore.dev/beta/errs"
)

func TestAdmin(t *testing.T) {
	tests := []struct {
		name  string
		input AdminParams
		want  error
	}{
		{
			name: "Test Admin Normal",
			input: AdminParams{
				Title:   "Test",
				StartAt: time.Now(),
				EndAt:   time.Now().Add(24 * time.Hour),
				Conditions: Condition{
					AgeStart: 18,
					AgeEnd:   30,
					Country:  []string{"TW", "JP"},
					Gender:   []string{"M", "F"},
				},
			},
			want: nil,
		},
		{
			name:  "Test Admin Error1",
			input: AdminParams{},
			want:  &errs.Error{Code: errs.InvalidArgument, Message: "Title, startAt and endAt are required"},
		},
		{
			name: "Test Admin Error2",
			input: AdminParams{
				Title:   "Test",
				StartAt: time.Now().Add(24 * time.Hour),
				EndAt:   time.Now(),
			},
			want: &errs.Error{Code: errs.InvalidArgument, Message: "StartAt must be before EndAt"},
		},
		{
			name: "Test Admin Error3",
			input: AdminParams{
				Title:   "Test",
				StartAt: time.Now(),
				EndAt:   time.Now().Add(24 * time.Hour),
				Conditions: Condition{
					AgeStart: 30,
					AgeEnd:   18,
				},
			},
			want: &errs.Error{Code: errs.InvalidArgument, Message: "AgeStart must be before AgeEnd"},
		},
		{
			name: "Test Admin Error4",
			input: AdminParams{
				Title:   "Test",
				StartAt: time.Now(),
				EndAt:   time.Now().Add(24 * time.Hour),
				Conditions: Condition{
					Gender: []string{"M", "F", "O"},
				},
			},
			want: &errs.Error{Code: errs.InvalidArgument, Message: "Invalid gender"},
		},
		{
			name: "Test Admin Error5",
			input: AdminParams{
				Title:   "Test",
				StartAt: time.Now(),
				EndAt:   time.Now().Add(24 * time.Hour),
				Conditions: Condition{
					Country: []string{"TW", "JP", "XX"},
				},
			},
			want: &errs.Error{Code: errs.InvalidArgument, Message: "Invalid country"},
		},
	}

	s, _ := initService()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s.Admin(context.Background(), tt.input)
			if got != nil {
				if got.Error() != tt.want.Error() {
					t.Errorf("Admin() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestPublic(t *testing.T) {
}
