package controllers

import (
	"fmt"
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
		c.JSON(http.StatusBadRequest, "Invalid request")
		return
	}

	if adminParams.Title == "" || adminParams.StartAt.IsZero() || adminParams.EndAt.IsZero() {
		c.JSON(http.StatusBadRequest, "Title, startAt and endAt are required")
		return
	}

	if adminParams.StartAt.After(adminParams.EndAt) {
		c.JSON(http.StatusBadRequest, "StartAt must be before EndAt")
		return
	}

	if adminParams.Conditions.AgeStart < 0 || adminParams.Conditions.AgeEnd > 100 || adminParams.Conditions.AgeStart > adminParams.Conditions.AgeEnd {
		c.JSON(http.StatusBadRequest, "Invalid age range")
		return
	}

	if (adminParams.Conditions.AgeStart == 0 || adminParams.Conditions.AgeEnd == 0) && adminParams.Conditions.AgeStart != adminParams.Conditions.AgeEnd {
		c.JSON(http.StatusBadRequest, "Invalid age range")
		return
	}

	if adminParams.Conditions.Gender != nil {
		for _, g := range adminParams.Conditions.Gender {
			if g != "M" && g != "F" {
				c.JSON(http.StatusBadRequest, "Invalid gender")
				return
			}
		}
	} else {
		adminParams.Conditions.Gender = []string{}
	}

	if adminParams.Conditions.Country != nil {
		for _, country := range adminParams.Conditions.Country {
			if countries.ByName(country) == countries.Unknown {
				c.JSON(http.StatusBadRequest, "Invalid country")
				return
			}
		}
	} else {
		adminParams.Conditions.Country = []string{}
	}

	if adminParams.Conditions.Platform != nil {
		for _, platform := range adminParams.Conditions.Platform {
			if platform != "ios" && platform != "android" && platform != "web" {
				c.JSON(http.StatusBadRequest, "Invalid platform")
				return
			}
		}
	} else {
		adminParams.Conditions.Platform = []string{}
	}

	models.CreateBanner(adminParams)

	c.JSON(http.StatusOK, "Banner created")
}

func SearchBanner(c *gin.Context) {
	var publicParams utils.PublicParams

	if err := c.ShouldBind(&publicParams); err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, "Invalid request")
		return
	}

	if publicParams.Age < 0 || publicParams.Age > 100 {
		c.JSON(http.StatusBadRequest, "Invalid age")
		return
	}

	if publicParams.Country != "" && countries.ByName(publicParams.Country) == countries.Unknown {
		c.JSON(http.StatusBadRequest, "Invalid country")
		return
	}

	if publicParams.Gender != "" && publicParams.Gender != "M" && publicParams.Gender != "F" {
		c.JSON(http.StatusBadRequest, "Invalid gender")
		return
	}

	if publicParams.Platform != "" && publicParams.Platform != "ios" && publicParams.Platform != "android" && publicParams.Platform != "web" {
		c.JSON(http.StatusBadRequest, "Invalid platform")
		return
	}

	if publicParams.Limit == 0 {
		publicParams.Limit = 5
	}

	item, err := models.SearchBanner(publicParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, "Internal server error")
		return
	}

	c.JSON(http.StatusOK, item)

}
