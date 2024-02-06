package ad

import (
	"context"
	"strings"
	"time"

	"encore.app/v1/utils"
	"encore.dev/beta/errs"
	"encore.dev/rlog"
	"encore.dev/storage/cache"
	"encore.dev/storage/sqldb"
	"github.com/biter777/countries"
	"github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//encore:service
type Service struct {
	db *gorm.DB
}

type AdminParams struct {
	Title      string    `json:"title"`
	StartAt    time.Time `json:"startAt"`
	EndAt      time.Time `json:"endAt"`
	Conditions Condition `json:"conditions"`
}

type PublicParams struct {
	Limit    int
	Offset   int
	Age      int
	Gender   string
	Country  string
	Platform string
}

type Item struct {
	Title string
	EndAt time.Time
}

type Condition struct {
	AgeStart int            `json:"ageStart"`
	AgeEnd   int            `json:"ageEnd"`
	Gender   pq.StringArray `json:"gender"`
	Country  pq.StringArray `json:"country"`
	Platform pq.StringArray `json:"platform"`
}

type AdResponse struct {
	Items []Item
}

type Banner struct {
	ID       uint
	Title    string
	StartAt  time.Time
	EndAt    time.Time
	AgeStart int
	AgeEnd   int
	Gender   pq.StringArray `gorm:"type:varchar[]"`
	Country  pq.StringArray `gorm:"type:varchar[]"`
	Platform pq.StringArray `gorm:"type:varchar[]"`
}

type ConditionCache struct {
	data []PublicParams
}

type SearchCache struct {
	data []Item
}

// for update cache data
var ConditionKeyspace = cache.NewStructKeyspace[string, ConditionCache](utils.Cluster, cache.KeyspaceConfig{
	KeyPattern:    "condition/:key",
	DefaultExpiry: cache.ExpireIn(5 * time.Minute),
})

// for search cache data
var SearchKeyspace = cache.NewStructKeyspace[PublicParams, SearchCache](utils.Cluster, cache.KeyspaceConfig{
	KeyPattern:    "search/:Limit/:Offset/:Age/:Gender/:Country/:Platform",
	DefaultExpiry: cache.ExpireIn(5 * time.Minute),
})

var adDB = sqldb.NewDatabase("api", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

func initService() (*Service, error) {

	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: adDB.Stdlib(),
	}))

	if err != nil {
		return nil, err
	}
	return &Service{db: db}, nil
}

func updateKeyspaceWhenCreate(ctx context.Context, kind string) {
	keys, err := ConditionKeyspace.Get(ctx, kind)
	if err != nil && !strings.Contains(err.Error(), cache.Miss.Error()) {
		rlog.Error("ERR", "err", err)
		return
	}

	for _, key := range keys.data {
		SearchKeyspace.Delete(ctx, key)
	}
}

func updateKeyspaceWhenRead(ctx context.Context, p PublicParams) {
	if p.Age > 0 {
		keys, err := ConditionKeyspace.Get(ctx, "age")
		if err != nil && !strings.Contains(err.Error(), cache.Miss.Error()) {
			rlog.Error("ERR", "err", err)
			return
		}

		keys.data = append(keys.data, p)
		ConditionKeyspace.Set(ctx, "age", keys)
	}

	if p.Country != "" {
		keys, err := ConditionKeyspace.Get(ctx, "country")
		if err != nil && !strings.Contains(err.Error(), cache.Miss.Error()) {
			rlog.Error("ERR", "err", err)
			return
		}

		keys.data = append(keys.data, p)
		ConditionKeyspace.Set(ctx, "country", keys)
	}

	if p.Gender != "" {
		keys, err := ConditionKeyspace.Get(ctx, "gender")
		if err != nil && !strings.Contains(err.Error(), cache.Miss.Error()) {
			rlog.Error("ERR", "err", err)
			return
		}

		keys.data = append(keys.data, p)
		ConditionKeyspace.Set(ctx, "gender", keys)
	}

	if p.Platform != "" {
		keys, err := ConditionKeyspace.Get(ctx, "platform")
		if err != nil && !strings.Contains(err.Error(), cache.Miss.Error()) {
			rlog.Error("ERR", "err", err)
			return
		}

		keys.data = append(keys.data, p)
		ConditionKeyspace.Set(ctx, "platform", keys)
	}
}

//encore:api public method=POST path=/api/v1/ad
func (s *Service) Admin(ctx context.Context, p AdminParams) error {
	if p.Title == "" || p.StartAt.IsZero() || p.EndAt.IsZero() {
		return &errs.Error{Code: errs.InvalidArgument, Message: "Title, startAt and endAt are required"}
	}

	if p.StartAt.After(p.EndAt) {
		return &errs.Error{Code: errs.InvalidArgument, Message: "StartAt must be before EndAt"}
	}

	if p.Conditions.AgeStart < 0 || p.Conditions.AgeEnd > 100 || p.Conditions.AgeStart > p.Conditions.AgeEnd {
		return &errs.Error{Code: errs.InvalidArgument, Message: "Invalid age range"}
	}

	if len(p.Conditions.Gender) > 0 {
		for _, g := range p.Conditions.Gender {
			if g != "M" && g != "F" {
				return &errs.Error{Code: errs.InvalidArgument, Message: "Invalid gender"}
			}
		}
	}

	if p.Conditions.Country != nil {
		for _, country := range p.Conditions.Country {
			if countries.ByName(country) == countries.Unknown {
				return &errs.Error{Code: errs.InvalidArgument, Message: "Invalid country"}
			}
		}
	}

	if p.Conditions.AgeStart > 0 && p.Conditions.AgeEnd < 100 {
		updateKeyspaceWhenCreate(ctx, "age")
	}

	if p.Conditions.Gender != nil {
		updateKeyspaceWhenCreate(ctx, "gender")
	}

	if p.Conditions.Country != nil {
		updateKeyspaceWhenCreate(ctx, "country")
	}

	if p.Conditions.Platform != nil {
		updateKeyspaceWhenCreate(ctx, "platform")
	}

	banner := Banner{
		Title:    p.Title,
		StartAt:  p.StartAt,
		EndAt:    p.EndAt,
		AgeStart: p.Conditions.AgeStart,
		AgeEnd:   p.Conditions.AgeEnd,
		Gender:   p.Conditions.Gender,
		Country:  p.Conditions.Country,
		Platform: p.Conditions.Platform,
	}

	if err := s.db.Table("banners").Create(&banner).Error; err != nil {
		return &errs.Error{Code: errs.Internal}
	}

	return nil
}

//encore:api public method=GET path=/api/v1/ad
func (s *Service) Public(ctx context.Context, p PublicParams) (*AdResponse, error) {
	rlog.Debug("DBG", "params", p)

	if p.Age < 0 || p.Age > 100 {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "Invalid age"}
	}

	if p.Country != "" && countries.ByName(p.Country) == countries.Unknown {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "Invalid country"}
	}

	if p.Gender != "" && p.Gender != "M" && p.Gender != "F" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "Invalid gender"}
	}

	if len(p.Platform) > 255 {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "Platform is too long"}
	}

	if p.Limit == 0 {
		p.Limit = 5
	}

	// get data from cache
	searchCache, err := SearchKeyspace.Get(ctx, p)
	if err != nil && !strings.Contains(err.Error(), cache.Miss.Error()) {
		rlog.Error("ERR", "err", err)
		return nil, &errs.Error{Code: errs.Internal}
	}

	items := []Item{}

	if len(searchCache.data) == 0 {
		query := "? BETWEEN start_at AND end_at"
		queryParams := []interface{}{time.Now()}
		if p.Age != 0 {
			query += " AND (? BETWEEN age_start AND age_end OR age_start IS NULL AND age_end IS NULL)"
			queryParams = append(queryParams, p.Age)
		}

		if p.Country != "" {
			query += " AND (? = ANY(country) OR country IS NULL)"
			queryParams = append(queryParams, p.Country)
		}

		if p.Gender != "" {
			query += " AND (? = ANY(gender) OR gender IS NULL)"
			queryParams = append(queryParams, p.Gender)
		}

		if p.Platform != "" {
			query += " AND (? = ANY(platform) OR platform IS NULL)"
			queryParams = append(queryParams, p.Platform)
		}

		var banners []Banner
		if err := s.db.Table("banners").Where(query, queryParams...).Limit(p.Limit).Offset(p.Offset).Find(&banners).Error; err != nil {
			return nil, &errs.Error{Code: errs.Internal}
		}

		if len(banners) != 0 {
			for _, banner := range banners {
				items = append(items, Item{Title: banner.Title, EndAt: banner.EndAt})
			}

			// set data to cache
			searchCache.data = items
			if err := SearchKeyspace.Set(ctx, p, searchCache); err != nil {
				rlog.Error("ERR", "err", err)
				return nil, &errs.Error{Code: errs.Internal}
			}

			updateKeyspaceWhenRead(ctx, p)
		}

	} else {
		rlog.Debug("DBG", "cache key", p, "data", searchCache.data)
		items = searchCache.data
	}

	return &AdResponse{Items: items}, nil
}
