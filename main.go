package main

import (
	"encoding/json"
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/timotheus90/go-playground/database"
	"github.com/timotheus90/go-playground/models"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

var (
	db                *database.Database
	err               error
	cleaningTasksPath = "/api/cleaning-tasks"
)

func getCleaningTasks(c echo.Context) error {
	cleaningTasks := []models.CleaningTask{{}}
	result := db.Model(&models.CleaningTask{}).Find(&cleaningTasks)
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, result.Error)
	}
	return c.JSON(http.StatusOK, cleaningTasks)
}

func createCleaningTask(c echo.Context) error {
	cleaningTask := models.CleaningTask{}
	err := json.NewDecoder(c.Request().Body).Decode(&cleaningTask)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	// create cleaning task in db
	result := db.Create(&cleaningTask)

	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, result.Error)
	}

	return c.JSON(http.StatusOK, cleaningTask)
}

func getCleaningTaskById(c echo.Context) error {
	id := c.Param("id")
	cleaningTask := models.CleaningTask{}
	result := db.Model(models.CleaningTask{}).First(&cleaningTask, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound)
		}

		return echo.NewHTTPError(http.StatusInternalServerError, result.Error)
	}

	return c.JSON(http.StatusOK, cleaningTask)
}

func updateCleaningTaskById(c echo.Context) error {
	cleaningTask := models.CleaningTask{}
	err := json.NewDecoder(c.Request().Body).Decode(&cleaningTask)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	id, _ := strconv.Atoi(c.Param("id"))
	result := db.First(&models.CleaningTask{}, id)
	if result.RowsAffected == 0 {
		return c.NoContent(http.StatusNotFound)
	}
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, result.Error)
	}

	// save updated cleaning task in db
	cleaningTask.ID = uint(id)
	// TODO: include id in request
	result = db.Save(&cleaningTask)
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, result.Error)
	}

	return c.JSON(http.StatusOK, cleaningTask)
}

func deleteCleaningTaskById(c echo.Context) error {
	id := c.Param("id")
	result := db.Find(&models.CleaningTask{}, id)
	if result.RowsAffected == 0 {
		return c.NoContent(http.StatusNotFound)
	}
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, result.Error)
	}

	result = db.Delete(&models.CleaningTask{}, id)
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, result.Error)
	}

	return c.NoContent(http.StatusNoContent)
}

func main() {
	// init database
	dataSourceName := "host=localhost user=postgres password=postgres port=5432 sslmode=disable"
	db, err = database.NewDatabase(dataSourceName)
	if err != nil {
		panic(err)
	}

	// enable debug logging
	err = db.AutoMigrate(&models.CleaningTask{})
	if err != nil {
		panic(err)
	}

	e := echo.New()

	cleaningTasksPath := "/api/cleaning-tasks"

	e.GET(cleaningTasksPath, getCleaningTasks)
	e.POST(cleaningTasksPath, createCleaningTask)
	// I dislike here that the handler methods has no explicit knowledge about the path (:id)
	// defining the path and handler together seems to be more explicit
	e.GET(cleaningTasksPath+"/:id", getCleaningTaskById)
	e.PUT(cleaningTasksPath+"/:id", updateCleaningTaskById)
	e.DELETE(cleaningTasksPath+"/:id", deleteCleaningTaskById)

	e.Logger.Fatal(e.Start(":1323"))
}
