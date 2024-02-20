package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"main/cache"
	"main/models"
	"main/routers"
	"main/tests/load_test"
	"main/utils"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redismock/v9"
	"github.com/joho/godotenv"
	"gotest.tools/assert"
)

var testRouter *gin.Engine

func prepareFilteringMockData() {
	banners := []models.Banner{
		{Title: "TestAge", StartAt: time.Now(), EndAt: time.Now().Add(1 * time.Hour), AgeStart: 18, AgeEnd: 30, Genders: []models.Gender{{Name: "F"}}, Countries: []models.Country{{Name: "TW"}, {Name: "JP"}}, Platforms: []models.Platform{{Name: "web"}}},
		{Title: "TestGender", StartAt: time.Now(), EndAt: time.Now().Add(2 * time.Hour), AgeStart: 31, AgeEnd: 40, Genders: []models.Gender{{Name: "M"}}, Countries: []models.Country{{Name: "TW"}, {Name: "JP"}}, Platforms: []models.Platform{{Name: "web"}}},
		{Title: "TestCountry", StartAt: time.Now(), EndAt: time.Now().Add(3 * time.Hour), AgeStart: 31, AgeEnd: 40, Genders: []models.Gender{{Name: "F"}}, Countries: []models.Country{{Name: "US"}, {Name: "UK"}}, Platforms: []models.Platform{{Name: "web"}}},
		{Title: "TestPlatform", StartAt: time.Now(), EndAt: time.Now().Add(4 * time.Hour), AgeStart: 31, AgeEnd: 40, Genders: []models.Gender{{Name: "F"}}, Countries: []models.Country{{Name: "TW"}, {Name: "JP"}}, Platforms: []models.Platform{{Name: "web"}, {Name: "android"}}},
		{Title: "TestAll", StartAt: time.Now(), EndAt: time.Now().Add(5 * time.Hour)},
	}
	load_test.DeleteAllData()
	for _, banner := range banners {
		models.DB.Create(&banner)
	}
}

func preparePaginationMockData() {
	banners := []models.Banner{
		{Title: "TestPagination1", StartAt: time.Now(), EndAt: time.Now().Add(1 * time.Hour), AgeStart: 31, AgeEnd: 40},
		{Title: "TestPagination2", StartAt: time.Now(), EndAt: time.Now().Add(2 * time.Hour), AgeStart: 31, AgeEnd: 40},
		{Title: "TestPagination3", StartAt: time.Now(), EndAt: time.Now().Add(3 * time.Hour), AgeStart: 31, AgeEnd: 40},
		{Title: "TestPagination4", StartAt: time.Now(), EndAt: time.Now().Add(4 * time.Hour), AgeStart: 31, AgeEnd: 40},
		{Title: "TestPagination5", StartAt: time.Now(), EndAt: time.Now().Add(5 * time.Hour), AgeStart: 31, AgeEnd: 40},
		{Title: "TestPagination6", StartAt: time.Now(), EndAt: time.Now().Add(6 * time.Hour), AgeStart: 31, AgeEnd: 40},
		{Title: "TestPagination7", StartAt: time.Now(), EndAt: time.Now().Add(7 * time.Hour), AgeStart: 31, AgeEnd: 40},
		{Title: "TestPagination8", StartAt: time.Now(), EndAt: time.Now().Add(8 * time.Hour), AgeStart: 31, AgeEnd: 40},
		{Title: "TestPagination9", StartAt: time.Now(), EndAt: time.Now().Add(9 * time.Hour), AgeStart: 31, AgeEnd: 40},
		{Title: "TestPagination10", StartAt: time.Now(), EndAt: time.Now().Add(10 * time.Hour), AgeStart: 31, AgeEnd: 40},
	}
	load_test.DeleteAllData()
	for _, banner := range banners {
		models.DB.Create(&banner)
	}
}

func TestMain(m *testing.M) {
	godotenv.Load("../../../.env")
	os.Setenv("APP_ENV", "test")
	fmt.Print(os.Getenv("APP_ENV"))
	models.Init()
	cache.Init()
	load_test.DeleteAllData()

	testRouter = routers.Init()
	m.Run()
}

func TestCreateBannerAPI(t *testing.T) {
	adminParams := utils.AdminParams{
		Title:   "test banner",
		StartAt: time.Now(),
		EndAt:   time.Now().Add(time.Duration(2) * time.Hour),
		Conditions: utils.ConditionParams{
			AgeStart: 20,
			AgeEnd:   30,
			Gender:   []string{"M"},
			Country:  []string{"TW"},
			Platform: []string{"ios"},
		},
	}
	jsonData, _ := json.Marshal(adminParams)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/ad", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestCreateBannerAPIClientError(t *testing.T) {
	tests := []struct {
		name string
		body utils.AdminParams
		want gin.H
	}{
		{
			name: "Empty title",
			body: utils.AdminParams{
				Title:   "",
				StartAt: time.Now(),
				EndAt:   time.Now().Add(time.Duration(2) * time.Hour),
			},
			want: gin.H{"error": "Title, startAt and endAt are required"},
		},
		{
			name: "Invalid time interval",
			body: utils.AdminParams{
				Title:   "test banner",
				StartAt: time.Now(),
				EndAt:   time.Now().Add(time.Duration(-2) * time.Hour),
			},
			want: gin.H{"error": "StartAt must be before EndAt"},
		},
		{
			name: "Invalid age range 1",
			body: utils.AdminParams{
				Title:   "test banner",
				StartAt: time.Now(),
				EndAt:   time.Now().Add(time.Duration(2) * time.Hour),
				Conditions: utils.ConditionParams{
					AgeStart: 30,
					AgeEnd:   20,
				},
			},
			want: gin.H{"error": "Invalid age range"},
		},
		{
			name: "Invalid age range 2",
			body: utils.AdminParams{
				Title:   "test banner",
				StartAt: time.Now(),
				EndAt:   time.Now().Add(time.Duration(2) * time.Hour),
				Conditions: utils.ConditionParams{
					AgeStart: 0,
					AgeEnd:   30,
				},
			},
			want: gin.H{"error": "Invalid age range"},
		},
		{
			name: "Invalid gender",
			body: utils.AdminParams{
				Title:   "test banner",
				StartAt: time.Now(),
				EndAt:   time.Now().Add(time.Duration(2) * time.Hour),
				Conditions: utils.ConditionParams{
					Gender: []string{"M", "F", "X"},
				},
			},
			want: gin.H{"error": "Invalid gender"},
		},
		{
			name: "Invalid country",
			body: utils.AdminParams{
				Title:   "test banner",
				StartAt: time.Now(),
				EndAt:   time.Now().Add(time.Duration(2) * time.Hour),
				Conditions: utils.ConditionParams{
					Country: []string{"X", "EVIL"},
				},
			},
			want: gin.H{"error": "Invalid country"},
		},
		{
			name: "Invalid platform",
			body: utils.AdminParams{
				Title:   "test banner",
				StartAt: time.Now(),
				EndAt:   time.Now().Add(time.Duration(2) * time.Hour),
				Conditions: utils.ConditionParams{
					Platform: []string{"ios", "android", "web", "X"},
				},
			},
			want: gin.H{"error": "Invalid platform"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.body)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/ad", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			testRouter.ServeHTTP(w, req)

			var got gin.H
			json.Unmarshal(w.Body.Bytes(), &got)
			assert.Equal(t, 400, w.Code)
			assert.Equal(t, tt.want["error"], got["error"])
		})
	}
}

func TestSearchBanners(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want []utils.Item
	}{
		{
			name: "Test Public Age",
			url:  "/api/v1/ad?age=20",
			want: []utils.Item{{Title: "TestAge"}, {Title: "TestAll"}},
		},
		{
			name: "Test Public Gender",
			url:  "/api/v1/ad?gender=M",
			want: []utils.Item{{Title: "TestGender"}, {Title: "TestAll"}},
		},
		{
			name: "Test Public Country",
			url:  "/api/v1/ad?country=US",
			want: []utils.Item{{Title: "TestCountry"}, {Title: "TestAll"}},
		},
		{
			name: "Test Public Platform",
			url:  "/api/v1/ad?platform=android",
			want: []utils.Item{{Title: "TestPlatform"}, {Title: "TestAll"}},
		},
		{
			name: "Test Public All",
			url:  "/api/v1/ad",
			want: []utils.Item{
				{Title: "TestAge"},
				{Title: "TestGender"},
				{Title: "TestCountry"},
				{Title: "TestPlatform"},
				{Title: "TestAll"},
			},
		},
		{
			name: "Test Public Pagination",
			url:  "/api/v1/ad?limit=6&offset=3",
			want: []utils.Item{
				{Title: "TestPagination4"},
				{Title: "TestPagination5"},
				{Title: "TestPagination6"},
				{Title: "TestPagination7"},
				{Title: "TestPagination8"},
				{Title: "TestPagination9"},
			},
		},
	}

	prepareFilteringMockData()

	for i, tt := range tests {
		if i == 5 {
			preparePaginationMockData()
		}

		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tt.url, nil)
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, 200, w.Code)

			var got []utils.Item
			json.Unmarshal(w.Body.Bytes(), &got)
			for i, item := range tt.want {
				assert.Assert(t, got[i].Title == item.Title, got[i].Title, i)

				if i < len(got)-1 {
					assert.Assert(t, got[i].EndAt.Before(got[i+1].EndAt))
				}
			}
		})
	}
}

func TestSearchBannersClientError(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want gin.H
	}{
		{
			name: "Invalid age",
			url:  "/api/v1/ad?age=101",
			want: gin.H{"error": "Invalid age"},
		},
		{
			name: "Invalid country",
			url:  "/api/v1/ad?country=XXX",
			want: gin.H{"error": "Invalid country"},
		},
		{
			name: "Invalid gender",
			url:  "/api/v1/ad?gender=X",
			want: gin.H{"error": "Invalid gender"},
		},
		{
			name: "Invalid platform",
			url:  "/api/v1/ad?platform=X",
			want: gin.H{"error": "Invalid platform"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tt.url, nil)
			testRouter.ServeHTTP(w, req)

			var got gin.H
			json.Unmarshal(w.Body.Bytes(), &got)

			assert.Equal(t, 400, w.Code)
			assert.Equal(t, tt.want["error"], got["error"])
		})
	}
}

func TestCacheMiddleware(t *testing.T) {
	rc, mock := redismock.NewClientMock()
	cache.RedisClient = rc

	prepareFilteringMockData()

	key := "/api/v1/ad?age=20"
	data := []utils.Item{{Title: "TestAge", EndAt: time.Now().Add(1 * time.Hour)}}
	jsonData, _ := json.Marshal(data)

	// Test cache hit
	mock.ExpectGet(key).SetVal(string(jsonData))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", key, nil)
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, string(jsonData), w.Body.String())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}

	// Test cache miss
	mock.ExpectGet(key).SetErr(fmt.Errorf("cache miss"))

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", key, nil)
	testRouter.ServeHTTP(w, req)

	var got []utils.Item
	json.Unmarshal(w.Body.Bytes(), &got)

	assert.Equal(t, 200, w.Code)

	want := []utils.Item{{Title: "TestAge"}, {Title: "TestAll"}}
	for i, item := range want {
		assert.Assert(t, got[i].Title == item.Title, got[i].Title, i)
	}
}
