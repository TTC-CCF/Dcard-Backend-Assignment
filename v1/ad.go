package ad

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"golang.org/x/sync/singleflight"

	"encore.dev/beta/errs"
	"encore.dev/storage/cache"
	"github.com/biter777/countries"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

//encore:service
type Service struct {
	db *gorm.DB
}

// AdminParams represents the input parameters for the Admin method.
type AdminParams struct {
	ContentType string          `header:"Content-Type"`
	Title       string          `json:"title"`
	StartAt     time.Time       `json:"startAt"`
	EndAt       time.Time       `json:"endAt"`
	Conditions  ConditionParams `json:"conditions"`
}

type ConditionParams struct {
	AgeStart int            `json:"ageStart"`
	AgeEnd   int            `json:"ageEnd"`
	Gender   pq.StringArray `json:"gender"`
	Country  pq.StringArray `json:"country"`
	Platform pq.StringArray `json:"platform"`
}

// PublicParams represents the input parameters for the Public method.
type PublicParams struct {
	Limit    int    `query:"limit"`
	Offset   int    `query:"offset"`
	Age      int    `query:"age"`
	Gender   string `query:"gender"`
	Country  string `query:"country"`
	Platform string `query:"platform"`
}

// PublicResponse represents the output parameters for the Public method.
type PublicResponse struct {
	Items []Item `json:"items"`
}

type Item struct {
	Title string    `json:"title"`
	EndAt time.Time `json:"endAt"`
}

var sfg singleflight.Group

var Cluster = cache.NewCluster("backend", cache.ClusterConfig{
	EvictionPolicy: cache.AllKeysLRU,
})

// key: age | gender | country | platform
// value: []PublicParams
var ConditionKeyspace = cache.NewStringKeyspace[string](Cluster, cache.KeyspaceConfig{
	KeyPattern:    "condition/:key",
	DefaultExpiry: cache.ExpireIn(5 * time.Minute),
})

// key: PublicParams
// value: []Item
var SearchKeyspace = cache.NewStringKeyspace[PublicParams](Cluster, cache.KeyspaceConfig{
	KeyPattern:    "search/:Limit/:Offset/:Age/:Gender/:Country/:Platform",
	DefaultExpiry: cache.ExpireIn(5 * time.Minute),
})

// encore will run this function on startup
func initService() (*Service, error) {
	db, err := InitDB()

	if err != nil {
		return nil, err
	}
	return &Service{db: db}, nil
}

func deleteKeyspaceWhenCreate(ctx context.Context, kind string) {
	// Delete cache when create new banner
	keys, err := ConditionKeyspace.Get(ctx, kind)
	if err != nil && !strings.Contains(err.Error(), cache.Miss.Error()) {
		return
	}

	var data []PublicParams
	json.Unmarshal([]byte(keys), &data)

	for _, key := range data {
		SearchKeyspace.Delete(ctx, key)
	}

	ConditionKeyspace.Delete(ctx, kind)
}

func updateKeyspaceWhenRead(ctx context.Context, kind string, p PublicParams) {
	// Append p to the keyspace given by kind
	keys, err := ConditionKeyspace.Get(ctx, kind)
	if err != nil && !strings.Contains(err.Error(), cache.Miss.Error()) {
		return
	}

	var data []PublicParams
	json.Unmarshal([]byte(keys), &data)

	data = append(data, p)

	jsonData, _ := json.Marshal(data)
	ConditionKeyspace.Set(ctx, kind, string(jsonData))
}

//encore:api public method=POST path=/api/v1/ad
func (s *Service) Admin(ctx context.Context, p AdminParams) error {
	// Validate the input parameters
	if p.ContentType != "application/json" {
		return &errs.Error{Code: errs.InvalidArgument, Message: "Invalid Content-Type"}
	}

	if p.Title == "" || p.StartAt.IsZero() || p.EndAt.IsZero() {
		return &errs.Error{Code: errs.InvalidArgument, Message: "Title, startAt and endAt are required"}
	}

	if p.StartAt.After(p.EndAt) {
		return &errs.Error{Code: errs.InvalidArgument, Message: "StartAt must be before EndAt"}
	}

	if p.Conditions.AgeStart < 0 || p.Conditions.AgeEnd > 100 || p.Conditions.AgeStart > p.Conditions.AgeEnd {
		return &errs.Error{Code: errs.InvalidArgument, Message: "Invalid age range"}
	}

	if (p.Conditions.AgeStart == 0 || p.Conditions.AgeEnd == 0) && p.Conditions.AgeStart != p.Conditions.AgeEnd {
		return &errs.Error{Code: errs.InvalidArgument, Message: "Invalid age range"}
	}

	if p.Conditions.Gender != nil {
		for _, g := range p.Conditions.Gender {
			if g != "M" && g != "F" {
				return &errs.Error{Code: errs.InvalidArgument, Message: "Invalid gender"}
			}
		}
	} else {
		p.Conditions.Gender = []string{}
	}

	if p.Conditions.Country != nil {
		for _, country := range p.Conditions.Country {
			if countries.ByName(country) == countries.Unknown {
				return &errs.Error{Code: errs.InvalidArgument, Message: "Invalid country"}
			}
		}
	} else {
		p.Conditions.Country = []string{}
	}

	if p.Conditions.Platform != nil {
		for _, platform := range p.Conditions.Platform {
			if platform != "ios" && platform != "android" && platform != "web" {
				return &errs.Error{Code: errs.InvalidArgument, Message: "Invalid platform"}
			}
		}
	} else {
		p.Conditions.Platform = []string{}
	}

	// Delete the corresponding cache when create new banner
	if p.Conditions.AgeStart > 0 && p.Conditions.AgeEnd < 100 {
		deleteKeyspaceWhenCreate(ctx, "age")
	}

	if p.Conditions.Gender != nil {
		deleteKeyspaceWhenCreate(ctx, "gender")
	}

	if p.Conditions.Country != nil {
		deleteKeyspaceWhenCreate(ctx, "country")
	}

	if p.Conditions.Platform != nil {
		deleteKeyspaceWhenCreate(ctx, "platform")
	}

	// Create new banner
	return s.CreateBanner(p)
}

//encore:api public method=GET path=/api/v1/ad
func (s *Service) Public(ctx context.Context, p PublicParams) (*PublicResponse, error) {
	// Validate the input parameters
	if p.Age < 0 || p.Age > 100 {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "Invalid age"}
	}

	if p.Country != "" && countries.ByName(p.Country) == countries.Unknown {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "Invalid country"}
	}

	if p.Gender != "" && p.Gender != "M" && p.Gender != "F" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "Invalid gender"}
	}

	if p.Platform != "" && p.Platform != "ios" && p.Platform != "android" && p.Platform != "web" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "Invalid platform"}
	}

	if p.Limit == 0 {
		p.Limit = 5
	}

	// Check if other request is already fetching the data from cache by singleflight group
	key, _ := json.Marshal(p)
	searchCache, err, _ := sfg.Do(string(key), func() (interface{}, error) {
		return SearchKeyspace.Get(ctx, p)
	})

	if err != nil && !strings.Contains(err.Error(), cache.Miss.Error()) {
		return nil, &errs.Error{Code: errs.Internal}
	}

	var data []Item
	json.Unmarshal([]byte(searchCache.(string)), &data)

	var items []Item

	// If cache miss, fetch data from database. Otherwise, return the data from cache
	if len(data) == 0 {
		// Also need singleflight group to prevent multiple requests to the database
		data, err, _ := sfg.Do(string(key), func() (interface{}, error) {
			return s.SearchBanners(p)
		})

		if err != nil {
			return nil, &errs.Error{Code: errs.Internal}
		}

		items = data.([]Item)

		if len(items) != 0 {
			// set data to cache
			jsonData, _ := json.Marshal(items)
			if err := SearchKeyspace.Set(ctx, p, string(jsonData)); err != nil {
				return nil, &errs.Error{Code: errs.Internal}
			}

			if p.Age != 0 {
				updateKeyspaceWhenRead(ctx, "age", p)
			}

			if p.Country != "" {
				updateKeyspaceWhenRead(ctx, "country", p)
			}

			if p.Gender != "" {
				updateKeyspaceWhenRead(ctx, "gender", p)
			}

			if p.Platform != "" {
				updateKeyspaceWhenRead(ctx, "platform", p)
			}
		} else {
			items = []Item{}
		}

	} else {
		items = data
	}

	return &PublicResponse{Items: items}, nil
}
