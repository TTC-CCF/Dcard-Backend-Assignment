package ad

import (
	"context"
	"os"
	"testing"
	"time"

	"encore.dev/beta/errs"
)

var s *Service

func prepareFilteringMockData() {
	banners := []Banner{
		{Title: "TestAge", StartAt: time.Now(), EndAt: time.Now().Add(24 * time.Hour), AgeStart: 18, AgeEnd: 30, Gender: []string{"F"}, Country: []string{"TW", "JP"}, Platform: []string{"web"}},
		{Title: "TestGender", StartAt: time.Now(), EndAt: time.Now().Add(24 * time.Hour), AgeStart: 31, AgeEnd: 40, Gender: []string{"M"}, Country: []string{"TW", "JP"}, Platform: []string{"web"}},
		{Title: "TestCountry", StartAt: time.Now(), EndAt: time.Now().Add(24 * time.Hour), AgeStart: 31, AgeEnd: 40, Gender: []string{"F"}, Country: []string{"US", "UK"}, Platform: []string{"web"}},
		{Title: "TestPlatform", StartAt: time.Now(), EndAt: time.Now().Add(24 * time.Hour), AgeStart: 31, AgeEnd: 40, Gender: []string{"F"}, Country: []string{"TW", "JP"}, Platform: []string{"ios", "android"}},
	}
	s.db.Exec("TRUNCATE banners")
	s.db.Table("banners").Create(&banners)
}

func preparePaginationMockData() {
	banners := []Banner{
		{Title: "TestPagination1", StartAt: time.Now(), EndAt: time.Now().Add(24 * time.Hour), AgeStart: 31, AgeEnd: 40},
		{Title: "TestPagination2", StartAt: time.Now(), EndAt: time.Now().Add(24 * time.Hour), AgeStart: 31, AgeEnd: 40},
		{Title: "TestPagination3", StartAt: time.Now(), EndAt: time.Now().Add(24 * time.Hour), AgeStart: 31, AgeEnd: 40},
		{Title: "TestPagination4", StartAt: time.Now(), EndAt: time.Now().Add(24 * time.Hour), AgeStart: 31, AgeEnd: 40},
		{Title: "TestPagination5", StartAt: time.Now(), EndAt: time.Now().Add(24 * time.Hour), AgeStart: 31, AgeEnd: 40},
		{Title: "TestPagination6", StartAt: time.Now(), EndAt: time.Now().Add(24 * time.Hour), AgeStart: 31, AgeEnd: 40},
		{Title: "TestPagination7", StartAt: time.Now(), EndAt: time.Now().Add(24 * time.Hour), AgeStart: 31, AgeEnd: 40},
		{Title: "TestPagination8", StartAt: time.Now(), EndAt: time.Now().Add(24 * time.Hour), AgeStart: 31, AgeEnd: 40},
		{Title: "TestPagination9", StartAt: time.Now(), EndAt: time.Now().Add(24 * time.Hour), AgeStart: 31, AgeEnd: 40},
		{Title: "TestPagination10", StartAt: time.Now(), EndAt: time.Now().Add(24 * time.Hour), AgeStart: 31, AgeEnd: 40},
	}
	s.db.Exec("TRUNCATE banners")
	s.db.Table("banners").Create(&banners)
}

func TestMain(m *testing.M) {
	s, _ = initService()
	os.Exit(m.Run())
}

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
			want: &errs.Error{Code: errs.InvalidArgument, Message: "Invalid age range"},
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
	tests := []struct {
		name  string
		input PublicParams
		want  *AdResponse
	}{
		{
			name: "Test Public Age",
			input: PublicParams{
				Age: 25,
			},
			want: &AdResponse{
				Items: []Item{{Title: "TestAge"}},
			},
		},
		{
			name: "Test Public Gender",
			input: PublicParams{
				Gender: "M",
			},
			want: &AdResponse{
				Items: []Item{{Title: "TestGender"}},
			},
		},
		{
			name: "Test Public Country",
			input: PublicParams{
				Country: "US",
			},
			want: &AdResponse{
				Items: []Item{{Title: "TestCountry"}},
			},
		},
		{
			name: "Test Public Platform",
			input: PublicParams{
				Platform: "ios",
			},
			want: &AdResponse{
				Items: []Item{{Title: "TestPlatform"}},
			},
		},
		{
			name: "Test Public Pagination",
			input: PublicParams{
				Limit:  4,
				Offset: 5,
			},
			want: &AdResponse{
				Items: []Item{
					{Title: "TestPagination6"},
					{Title: "TestPagination7"},
					{Title: "TestPagination8"},
					{Title: "TestPagination9"},
				},
			},
		},
	}

	for i, tt := range tests {
		if i == 0 {
			prepareFilteringMockData()
		}
		if i == 4 {
			preparePaginationMockData()
		}
		t.Run(tt.name, func(t *testing.T) {
			got, _ := s.Public(context.Background(), tt.input)
			if got != nil {
				for j := range got.Items {
					if got.Items[j].Title != tt.want.Items[j].Title {
						t.Errorf("Public() = %v, want %v", got, tt.want)
					}
					if got.Items[j].EndAt.Before(time.Now()) {
						t.Errorf("Public() = %v, want %v", got, tt.want)
					}
				}
			}

		})
	}

}
