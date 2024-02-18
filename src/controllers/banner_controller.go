package controllers

import (
	"fmt"
	"main/cache"
	"main/models"
	"main/utils"
	"net/http"

	"github.com/biter777/countries"
	"github.com/gin-gonic/gin"
)

func CreateBanner(c *gin.Context) {
	var adminParams utils.AdminParams

	if err := c.Bind(&adminParams); err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if adminParams.Title == "" || adminParams.StartAt.IsZero() || adminParams.EndAt.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title, startAt and endAt are required"})
		return
	}

	if adminParams.StartAt.After(adminParams.EndAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "StartAt must be before EndAt"})
		return
	}

	if adminParams.Conditions.AgeStart < 0 || adminParams.Conditions.AgeEnd > 100 || adminParams.Conditions.AgeStart > adminParams.Conditions.AgeEnd {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid age range"})
		return
	}

	if (adminParams.Conditions.AgeStart == 0 || adminParams.Conditions.AgeEnd == 0) && adminParams.Conditions.AgeStart != adminParams.Conditions.AgeEnd {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid age range"})
		return
	}

	if adminParams.Conditions.Gender != nil {
		for _, g := range adminParams.Conditions.Gender {
			if g != "M" && g != "F" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid gender"})
				return
			}
		}
	} else {
		adminParams.Conditions.Gender = []string{}
	}

	if adminParams.Conditions.Country != nil {
		for _, country := range adminParams.Conditions.Country {
			if countries.ByName(country) == countries.Unknown {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid country"})
				return
			}
		}
	} else {
		adminParams.Conditions.Country = []string{}
	}

	if adminParams.Conditions.Platform != nil {
		for _, platform := range adminParams.Conditions.Platform {
			if platform != "ios" && platform != "android" && platform != "web" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid platform"})
				return
			}
		}
	} else {
		adminParams.Conditions.Platform = []string{}
	}

	if err := models.CreateBanner(adminParams); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// delete related cache
	if adminParams.Conditions.AgeStart != 0 {
		cache.DeleteConditionCache(c, "age")
	}
	if len(adminParams.Conditions.Country) != 0 {
		cache.DeleteConditionCache(c, "country")
	}
	if len(adminParams.Conditions.Gender) != 0 {
		cache.DeleteConditionCache(c, "gender")
	}
	if len(adminParams.Conditions.Platform) != 0 {
		cache.DeleteConditionCache(c, "platform")
	}

	c.JSON(http.StatusOK, gin.H{"message": "Banner created"})
}

func SearchBanners(c *gin.Context) {
	var publicParams utils.PublicParams
	if err := c.ShouldBind(&publicParams); err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if publicParams.Age < 0 || publicParams.Age > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid age"})
		return
	}

	if publicParams.Country != "" && countries.ByName(publicParams.Country) == countries.Unknown {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid country"})
		return
	}

	if publicParams.Gender != "" && publicParams.Gender != "M" && publicParams.Gender != "F" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid gender"})
		return
	}

	if publicParams.Platform != "" && publicParams.Platform != "ios" && publicParams.Platform != "android" && publicParams.Platform != "web" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid platform"})
		return
	}

	if publicParams.Limit == 0 {
		publicParams.Limit = 5
	}

	// single flight
	key := c.Request.URL.Path + "?" + c.Request.URL.RawQuery
	data, err, _ := utils.Sfg.Do(key, func() (interface{}, error) {
		return models.SearchBanner(publicParams)
	})

	item := data.([]utils.Item)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	cache.SetCache(c, key, item)

	if publicParams.Age != 0 {
		cache.AddConditionCache(c, "age", key)
	}
	if publicParams.Country != "" {
		cache.AddConditionCache(c, "country", key)
	}
	if publicParams.Gender != "" {
		cache.AddConditionCache(c, "gender", key)
	}
	if publicParams.Platform != "" {
		cache.AddConditionCache(c, "platform", key)
	}

	c.JSON(http.StatusOK, item)
}
