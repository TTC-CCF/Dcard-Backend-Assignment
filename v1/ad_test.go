package ad

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"encore.dev/beta/errs"
	"encore.dev/storage/cache"
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

func TestDeleteKeyspaceWhenCreate(t *testing.T) {
	param := []PublicParams{
		{Limit: 10, Offset: 3, Age: 25, Gender: "M", Country: "TW", Platform: "web"},
		{Age: 25, Gender: "M"},
		{Age: 30, Country: "JP"},
		{Age: 10, Platform: "ios"},
	}
	jsonData, _ := json.Marshal(param)
	ConditionKeyspace.Set(context.Background(), "age", string(jsonData))

	for _, p := range param {
		SearchKeyspace.Set(context.Background(), p, "data that queried from database")
	}

	deleteKeyspaceWhenCreate(context.Background(), "age")

	_, err := ConditionKeyspace.Get(context.Background(), "age")
	if !strings.Contains(err.Error(), cache.Miss.Error()) {
		t.Error("deleteKeyspaceWhenCreate() not delete keyspace")
	}

	for _, p := range param {
		_, err := SearchKeyspace.Get(context.Background(), p)
		if !strings.Contains(err.Error(), cache.Miss.Error()) {
			t.Error("deleteKeyspaceWhenCreate() not delete keyspace")
		}
	}
}

func TestUpdateKeyspaceWhenRead(t *testing.T) {
	param := []PublicParams{
		{Limit: 10, Offset: 3, Age: 25, Gender: "M", Country: "TW", Platform: "web"},
		{Age: 25, Gender: "M"},
		{Age: 30, Country: "JP"},
		{Age: 10, Platform: "ios"},
	}
	jsonData, _ := json.Marshal(param)
	ConditionKeyspace.Set(context.Background(), "age", string(jsonData))

	newParam := PublicParams{
		Age:    25,
		Limit:  100,
		Offset: 40,
	}

	updateKeyspaceWhenRead(context.Background(), "age", newParam)

	var got []PublicParams
	data, _ := ConditionKeyspace.Get(context.Background(), "age")
	json.Unmarshal([]byte(data), &got)

	want := append(param, newParam)

	for i := range got {
		if got[i] != want[i] {
			t.Errorf("updateKeyspaceWhenRead() = %v, want %v", got, want)
		}
	}
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
				ContentType: "application/json",
				Title:       "Test",
				StartAt:     time.Now(),
				EndAt:       time.Now().Add(24 * time.Hour),
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
			name: "Test Admin Argument Error",
			input: AdminParams{
				ContentType: "application/json",
			},
			want: &errs.Error{Code: errs.InvalidArgument, Message: "Title, startAt and endAt are required"},
		},
		{
			name: "Test Admin Time Error",
			input: AdminParams{
				ContentType: "application/json",
				Title:       "Test",
				StartAt:     time.Now().Add(24 * time.Hour),
				EndAt:       time.Now(),
			},
			want: &errs.Error{Code: errs.InvalidArgument, Message: "StartAt must be before EndAt"},
		},
		{
			name: "Test Admin Age Error",
			input: AdminParams{
				ContentType: "application/json",
				Title:       "Test",
				StartAt:     time.Now(),
				EndAt:       time.Now().Add(24 * time.Hour),
				Conditions: Condition{
					AgeStart: 30,
					AgeEnd:   18,
				},
			},
			want: &errs.Error{Code: errs.InvalidArgument, Message: "Invalid age range"},
		},
		{
			name: "Test Admin Gender Error",
			input: AdminParams{
				ContentType: "application/json",
				Title:       "Test",
				StartAt:     time.Now(),
				EndAt:       time.Now().Add(24 * time.Hour),
				Conditions: Condition{
					Gender: []string{"M", "F", "O"},
				},
			},
			want: &errs.Error{Code: errs.InvalidArgument, Message: "Invalid gender"},
		},
		{
			name: "Test Admin Country Error",
			input: AdminParams{
				ContentType: "application/json",
				Title:       "Test",
				StartAt:     time.Now(),
				EndAt:       time.Now().Add(24 * time.Hour),
				Conditions: Condition{
					Country: []string{"TW", "JP", "XX"},
				},
			},
			want: &errs.Error{Code: errs.InvalidArgument, Message: "Invalid country"},
		},
		{
			name: "Test Admin Platform Error",
			input: AdminParams{
				ContentType: "application/json",
				Title:       "Test",
				StartAt:     time.Now(),
				EndAt:       time.Now().Add(24 * time.Hour),
				Conditions: Condition{
					Platform: []string{"web", "ios", "android", "evil"},
				},
			},
			want: &errs.Error{Code: errs.InvalidArgument, Message: "Invalid platform"},
		},
		{
			name: "Test Admin Content-Type Error",
			input: AdminParams{
				ContentType: "text/html",
				Title:       "Test",
				StartAt:     time.Now(),
				EndAt:       time.Now().Add(24 * time.Hour),
			},
			want: &errs.Error{Code: errs.InvalidArgument, Message: "Invalid Content-Type"},
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
		name    string
		input   PublicParams
		want    *PublicResponse
		wantErr error
	}{
		{
			name: "Test Public Age",
			input: PublicParams{
				Age: 25,
			},
			want: &PublicResponse{
				Items: []Item{{Title: "TestAge"}},
			},
		},
		{
			name: "Test Public Gender",
			input: PublicParams{
				Gender: "M",
			},
			want: &PublicResponse{
				Items: []Item{{Title: "TestGender"}},
			},
		},
		{
			name: "Test Public Country",
			input: PublicParams{
				Country: "US",
			},
			want: &PublicResponse{
				Items: []Item{{Title: "TestCountry"}},
			},
		},
		{
			name: "Test Public Platform",
			input: PublicParams{
				Platform: "ios",
			},
			want: &PublicResponse{
				Items: []Item{{Title: "TestPlatform"}},
			},
		},
		{
			name: "Test Public Pagination",
			input: PublicParams{
				Limit:  4,
				Offset: 5,
			},
			want: &PublicResponse{
				Items: []Item{
					{Title: "TestPagination6"},
					{Title: "TestPagination7"},
					{Title: "TestPagination8"},
					{Title: "TestPagination9"},
				},
			},
		},
		{
			name: "Test Public Age Error",
			input: PublicParams{
				Age: 101,
			},
			wantErr: &errs.Error{Code: errs.InvalidArgument, Message: "Invalid age"},
		},
		{
			name: "Test Public Country Error",
			input: PublicParams{
				Country: "EVIL",
			},
			wantErr: &errs.Error{Code: errs.InvalidArgument, Message: "Invalid country"},
		},
		{
			name: "Test Public Gender Error",
			input: PublicParams{
				Gender: "EVIL",
			},
			wantErr: &errs.Error{Code: errs.InvalidArgument, Message: "Invalid gender"},
		},
		{
			name: "Test Public Platform Error",
			input: PublicParams{
				Platform: "EVIL",
			},
			wantErr: &errs.Error{Code: errs.InvalidArgument, Message: "Invalid platform"},
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
			got, err := s.Public(context.Background(), tt.input)
			if tt.want != nil {
				for j := range got.Items {
					if got.Items[j].Title != tt.want.Items[j].Title {
						t.Errorf("Public() = %v, want %v", got, tt.want)
					}
					if got.Items[j].EndAt.Before(time.Now()) {
						t.Errorf("Public() = %v, want %v", got, tt.want)
					}
				}
			}
			if tt.wantErr != nil {
				if err.Error() != tt.wantErr.Error() {
					t.Errorf("Public() = %v, want %v", got, tt.want)
				}
			}

		})
	}
}
