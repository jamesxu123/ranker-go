package routes

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"ranker-go/lib"
	"ranker-go/schema"
)

func startSchedulerHandler(c echo.Context) error {
	var competitors []schema.Competitor
	result := schema.DB.Find(&competitors)
	if result.Error != nil {
		return c.String(http.StatusInternalServerError, "database error")
	}
	err := lib.SeedStart(competitors, 3)
	if err != nil {
		if err.Error() == "settings:scheduler-state has changed" {
			return c.String(http.StatusConflict, "scheduler-state has changed")
		} else {
			return c.String(http.StatusInternalServerError, err.Error())
		}
	}
	return c.String(http.StatusOK, "matches scheduled")
}

func resetScheduler(c echo.Context) error {
	lib.SetSchedulerState(schema.StateNone)
	return c.String(http.StatusOK, "scheduler reset")
}

func getAllMatchesHandler(c echo.Context) error {
	matches, err := lib.GetAllMatches()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, matches)
}

func deleteAllMatchesHandler(c echo.Context) error {
	err := lib.DeleteAllMatches()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.String(http.StatusOK, "deleted all matches")
}

func AddMatchScheduler(group *echo.Group) {
	group.POST("/start", startSchedulerHandler)
	group.PUT("/state/reset", resetScheduler)
	group.GET("/matches/all", getAllMatchesHandler)
	group.DELETE("/matches/all", deleteAllMatchesHandler)
}
