package main

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"ranker-go/glicko_go"
	"ranker-go/routes"
	"ranker-go/schema"
)

func test() {
	fmt.Println("Hello, World!")
	p1 := glicko_go.Glicko2From1(glicko_go.Glicko1{
		Rating: 1500,
		Sigma:  0.06,
		Rd:     200,
	})

	o1 := glicko_go.Glicko2From1(glicko_go.Glicko1{
		Rating: 1400,
		Sigma:  0.06,
		Rd:     30,
	})

	o2 := glicko_go.Glicko2From1(glicko_go.Glicko1{
		Rating: 1550,
		Sigma:  0.06,
		Rd:     100,
	})

	o3 := glicko_go.Glicko2From1(glicko_go.Glicko1{
		Rating: 1700,
		Sigma:  0.06,
		Rd:     300,
	})

	scores := []float64{glicko_go.WIN, glicko_go.LOSE, glicko_go.LOSE}
	opps := []glicko_go.Glicko2{o1, o2, o3}

	pF, err := p1.ProcessMatches(opps, scores)

	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("mu: %f, sigma %f, phi: %f\n", pF.Mu, pF.Sigma, pF.Phi)

	p1AsG1 := glicko_go.Glicko1From2(pF)
	fmt.Printf("Rating: %f, sigma %f, RD: %f\n", p1AsG1.Rating, p1AsG1.Sigma, p1AsG1.Rd)
}

func main() {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World")
	})
	competitorGroup := e.Group("/competitor")
	routes.AddCompetitorRoutes(competitorGroup)
	err := schema.Open()
	if err != nil {
		panic("failed to open database connection")
	}

	redisErr := schema.RedisOpen("redis://localhost:6379")
	if redisErr != nil {
		panic(redisErr)
	}

	e.Logger.Fatal(e.Start(":2997"))
}
