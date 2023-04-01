package main

import (
	"github.com/labstack/echo/v4"
	"github.com/timotheus90/go-playground/controller"
	"github.com/timotheus90/go-playground/database"
)

func main() {
	dataSourceName := "host=localhost user=postgres password=postgres port=5432 sslmode=disable"
	db, err := database.NewDatabase(dataSourceName)
	if err != nil {
		panic(err)
	}

	cleaningTaskRepository := database.CleaningTaskRepository{DB: db}

	e := echo.New()
	controller.SetupCleaningTaskEndpoints(e, &cleaningTaskRepository)
	e.Logger.Fatal(e.Start(":1323"))
}
