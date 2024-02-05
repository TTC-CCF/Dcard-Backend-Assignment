package ad

import (
	"context"
	"time"

	"encore.dev/beta/errs"
	"encore.dev/rlog"
	"encore.dev/storage/sqldb"
	"github.com/biter777/countries"
	"github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var adDB = sqldb.NewDatabase("api", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

//encore:service
type Service struct {
	db *gorm.DB
}

func initService() (*Service, error) {
	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: adDB.Stdlib(),
	}))
	if err != nil {
		return nil, err
	}
	return &Service{db: db}, nil
}

//encore:api public method=POST path=/api/v1/ad
func (s *Service) Admin(ctx context.Context, p AdminParams) error {
	if p.Title == "" || p.StartAt.IsZero() || p.EndAt.IsZero() {
		return &errs.Error{Code: errs.InvalidArgument, Message: "Title, startAt and endAt are required"}
	}

	if p.StartAt.After(p.EndAt) {
		return &errs.Error{Code: errs.InvalidArgument, Message: "StartAt must be before EndAt"}
	}

	if p.Conditions.AgeStart > p.Conditions.AgeEnd {
		return &errs.Error{Code: errs.InvalidArgument, Message: "AgeStart must be before AgeEnd"}
	}

	if len(p.Conditions.Gender) > 2 {
		return &errs.Error{Code: errs.InvalidArgument, Message: "Invalid gender"}
	}

	if p.Conditions.Country != nil {
		for _, country := range p.Conditions.Country {
			if countries.ByName(country) == countries.Unknown {
				return &errs.Error{Code: errs.InvalidArgument, Message: "Invalid country"}
			}
		}
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

	if p.Offset == 0 {
		p.Offset = 0
	}
	rlog.Debug("DBG", "params", p)
	query := "start_at <= ? AND end_at >= ? "
	queryParams := []interface{}{time.Now(), time.Now()}
	if p.Age != 0 {
		query += "AND age_start <= ? AND age_end >= ? "
		queryParams = append(queryParams, p.Age, p.Age)
	}

	if p.Country != "" {
		query += "AND ? = ANY(country) "
		queryParams = append(queryParams, p.Country)
	}

	if p.Gender != "" {
		query += "AND ? = ANY(gender) "
		queryParams = append(queryParams, p.Gender)
	}

	if p.Platform != "" {
		query += "AND ? = ANY(platform) "
		queryParams = append(queryParams, p.Platform)
	}

	rlog.Debug("DBG", "query", query, "queryParams", queryParams)

	var banners []Banner
	if err := s.db.Table("banners").Where(query, queryParams...).Limit(p.Limit).Offset(p.Offset).Find(&banners).Error; err != nil {
		return nil, &errs.Error{Code: errs.Internal}
	}

	var items []struct {
		Title string
		EndAt time.Time
	}

	for _, banner := range banners {
		items = append(items, struct {
			Title string
			EndAt time.Time
		}{Title: banner.Title, EndAt: banner.EndAt})
	}

	return &AdResponse{Items: items}, nil
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

type Condition struct {
	AgeStart int            `json:"ageStart"`
	AgeEnd   int            `json:"ageEnd"`
	Gender   pq.StringArray `json:"gender"`
	Country  pq.StringArray `json:"country"`
	Platform pq.StringArray `json:"platform"`
}

type AdResponse struct {
	Message string
	Items   []struct {
		Title string
		EndAt time.Time
	}
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
