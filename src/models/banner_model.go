package models

import (
	"fmt"
	"main/utils"
	"time"

	"gorm.io/gorm"
)

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
	Name string `gorm:"unique"`
}

type Country struct {
	ID   uint
	Name string `gorm:"unique"`
}

type Platform struct {
	ID   uint
	Name string `gorm:"unique"`
}

func (g *Gender) BeforeCreate(tx *gorm.DB) (err error) {
	var dup Gender
	if result := tx.First(&dup, "name = ?", g.Name); result.RowsAffected != 0 {
		g.ID = dup.ID
		return nil
	}
	return nil
}

func (c *Country) BeforeCreate(tx *gorm.DB) (err error) {
	var dup Country
	if result := tx.First(&dup, "name = ?", c.Name); result.RowsAffected != 0 {
		c.ID = dup.ID
		return nil
	}
	return nil
}

func (p *Platform) BeforeCreate(tx *gorm.DB) (err error) {
	var dup Platform
	if result := tx.First(&dup, "name = ?", p.Name); result.RowsAffected != 0 {
		p.ID = dup.ID
		return nil
	}
	return nil
}

func CreateBanner(p utils.AdminParams) error {
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

	return DB.Create(&banner).Error
}

func SearchBanner(p utils.PublicParams) ([]utils.Item, error) {
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
	res := DB.
		Distinct("banners.id, banners.title, banners.end_at").
		Joins("LEFT OUTER JOIN banner_gender ON banners.id = banner_gender.banner_id").
		Joins("LEFT OUTER JOIN genders ON genders.id = banner_gender.gender_id").
		Joins("LEFT OUTER JOIN banner_country ON banners.id = banner_country.banner_id").
		Joins("LEFT OUTER JOIN countries ON countries.id = banner_country.country_id").
		Joins("LEFT OUTER JOIN banner_platform ON banners.id = banner_platform.banner_id").
		Joins("LEFT OUTER JOIN platforms ON platforms.id = banner_platform.platform_id").
		Where(query, queryParams...).Order("end_at asc").Find(&banners)

	err := res.Error
	if err != nil {
		return nil, err
	}

	var items []utils.Item
	for i, b := range banners {
		if i >= p.Offset && i < p.Offset+p.Limit {
			items = append(items, utils.Item{Title: b.Title, EndAt: b.EndAt})
		}
	}
	fmt.Println(items)
	return items, nil
}
