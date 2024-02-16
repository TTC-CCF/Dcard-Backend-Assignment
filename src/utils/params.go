package utils

import "time"

type AdminParams struct {
	Title      string          `form:"title"`
	StartAt    time.Time       `form:"startAt"`
	EndAt      time.Time       `form:"endAt"`
	Conditions ConditionParams `form:"conditions"`
}

type ConditionParams struct {
	AgeStart int      `form:"ageStart"`
	AgeEnd   int      `form:"ageEnd"`
	Gender   []string `form:"gender"`
	Country  []string `form:"country"`
	Platform []string `form:"platform"`
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
