package routes

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"ranker-go/schema"
)

func addCompetitor(c echo.Context) error {
	var newCompetitor schema.Competitor
	err := c.Bind(&newCompetitor)
	if err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}
	fmt.Printf("Name: %s, Location: %s, Description: %s\n", newCompetitor.Name, newCompetitor.Location, newCompetitor.Description)
	createErr := newCompetitor.CreateInDb()
	if createErr != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}
	return c.String(http.StatusOK, "user created")
}

func getAllCompetitors(c echo.Context) error {
	var competitors []schema.Competitor
	result := schema.DB.Find(&competitors)
	if result.Error != nil {
		return c.String(http.StatusInternalServerError, "database error")
	}
	return c.JSON(http.StatusOK, competitors)
}

func AddCompetitorRoutes(group *echo.Group) {
	group.POST("/", addCompetitor)
	group.GET("/", getAllCompetitors)
}
