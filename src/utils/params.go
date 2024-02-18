package utils

import "time"

type AdminParams struct {
	Title      string          `form:"title" json:"title"`
	StartAt    time.Time       `form:"startAt" json:"startAt"`
	EndAt      time.Time       `form:"endAt" json:"endAt"`
	Conditions ConditionParams `form:"conditions" json:"conditions"`
}

type ConditionParams struct {
	AgeStart int      `form:"ageStart" json:"ageStart"`
	AgeEnd   int      `form:"ageEnd" json:"ageEnd"`
	Gender   []string `form:"gender" json:"gender"`
	Country  []string `form:"country" json:"country"`
	Platform []string `form:"platform" json:"platform"`
}

type PublicParams struct {
	Limit    int    `form:"limit"`
	Offset   int    `form:"offset"`
	Age      int    `form:"age"`
	Gender   string `form:"gender"`
	Country  string `form:"country"`
	Platform string `form:"platform"`
}

type Item struct {
	Title string    `json:"title"`
	EndAt time.Time `json:"endAt"`
}

type CachedItem struct {
	Data []Item `json:"data"`
}
