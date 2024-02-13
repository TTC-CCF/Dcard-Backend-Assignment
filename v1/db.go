package ad

import (
	"time"

	"encore.dev/storage/sqldb"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Schemas
type Banner struct {
	ID        uint
	Title     string
	StartAt   time.Time
	EndAt     time.Time
	AgeStart  int
	AgeEnd    int
	Genders   []Gender   `gorm:"many2many:banner_gender;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Countries []Country  `gorm:"many2many:banner_country;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Platforms []Platform `gorm:"many2many:banner_platform;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type Gender struct {
	ID   uint
	Name string
}

type Country struct {
	ID   uint
	Name string
}

type Platform struct {
	ID   uint
	Name string
}

var adDB = sqldb.NewDatabase("api", sqldb.DatabaseConfig{Migrations: "./"})

func InitDB() (*gorm.DB, error) {
	db, _ := gorm.Open(postgres.New(postgres.Config{
		Conn: adDB.Stdlib(),
	}))

	db.AutoMigrate(&Banner{}, &Gender{}, &Country{}, &Platform{})
	return db, nil
}

func (s *Service) CreateBanner(p AdminParams) error {
	var genders []Gender
	var countries []Country
	var platforms []Platform

	for _, g := range p.Conditions.Gender {
		genders = append(genders, Gender{Name: g})
	}

	for _, c := range p.Conditions.Country {
		countries = append(countries, Country{Name: c})
	}

	for _, p := range p.Conditions.Platform {
		platforms = append(platforms, Platform{Name: p})
	}

	banner := Banner{
		Title:     p.Title,
		StartAt:   p.StartAt,
		EndAt:     p.EndAt,
		AgeStart:  p.Conditions.AgeStart,
		AgeEnd:    p.Conditions.AgeEnd,
		Genders:   genders,
		Countries: countries,
		Platforms: platforms,
	}

	return s.db.Create(&banner).Error
}

func (s *Service) SearchBanners(p PublicParams) ([]Item, error) {
	var banners []Banner
	query := "NOW() BETWEEN start_at AND end_at"
	queryParams := []interface{}{}

	if p.Age != 0 {
		query += " AND (? BETWEEN age_start AND age_end OR age_end = 0 AND age_start = 0)"
		queryParams = append(queryParams, p.Age)
	}

	if p.Country != "" {
		query += " AND (countries.name = ? OR countries.name IS NULL)"
		queryParams = append(queryParams, p.Country)
	}

	if p.Gender != "" {
		query += " AND (genders.name = ? OR genders.name IS NULL)"
		queryParams = append(queryParams, p.Gender)
	}

	if p.Platform != "" {
		query += " AND (platforms.name = ? OR platforms.name IS NULL)"
		queryParams = append(queryParams, p.Platform)
	}

	res := s.db.
		Distinct("banners.id, banners.title, banners.end_at").
		Joins("JOIN banner_gender ON banners.id = banner_gender.banner_id").
		Joins("JOIN genders ON genders.id = banner_gender.gender_id").
		Joins("JOIN banner_country ON banners.id = banner_country.banner_id").
		Joins("JOIN countries ON countries.id = banner_country.country_id").
		Joins("JOIN banner_platform ON banners.id = banner_platform.banner_id").
		Joins("JOIN platforms ON platforms.id = banner_platform.platform_id").
		Where(query, queryParams...).Limit(p.Limit).Offset(p.Offset).Find(&banners)

	err := res.Error
	if err != nil {
		return nil, err
	}

	var items []Item
	for _, b := range banners {
		items = append(items, Item{Title: b.Title, EndAt: b.EndAt})
	}
	return items, nil
}
