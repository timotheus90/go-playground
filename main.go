package main

import (
	"encoding/json"
	"errors"
	"github.com/labstack/echo/v4"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"net/http"
	"strconv"
	"time"
)

type CleaningTask struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"-"`
	UpdatedAt   time.Time      `json:"-"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Description string         `json:"description"`
	DueDate     time.Time      `json:"dueDate"`
	Assignee    string         `json:"assignee"`
	Completed   bool           `json:"completed"`
	Category    TaskCategory   `json:"category"`
}

type TaskCategory string

const (
	CategoryKitchen TaskCategory = "kitchen"
	CategoryBaths   TaskCategory = "baths"
	CategoryFloors  TaskCategory = "floors"
	CategoryOther   TaskCategory = "other"
)

var (
	db  *gorm.DB
	err error
)

func getCleaningTasks(c echo.Context) error {
	cleaningTasks := []CleaningTask{{}}
	result := db.Model(&CleaningTask{}).Find(&cleaningTasks)
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, result.Error)
	}
	return c.JSON(http.StatusOK, cleaningTasks)
}

func createCleaningTask(c echo.Context) error {
	cleaningTask := CleaningTask{}
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
	cleaningTask := CleaningTask{}
	result := db.Model(CleaningTask{}).First(&cleaningTask, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound)
		}

		return echo.NewHTTPError(http.StatusInternalServerError, result.Error)
	}

	return c.JSON(http.StatusOK, cleaningTask)
}

func updateCleaningTaskById(c echo.Context) error {
	cleaningTask := CleaningTask{}
	err := json.NewDecoder(c.Request().Body).Decode(&cleaningTask)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	id, _ := strconv.Atoi(c.Param("id"))
	result := db.First(&CleaningTask{}, id)
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
	result := db.Find(&CleaningTask{}, id)
	if result.RowsAffected == 0 {
		return c.NoContent(http.StatusNotFound)
	}
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, result.Error)
	}

	result = db.Delete(&CleaningTask{}, id)
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, result.Error)
	}

	return c.NoContent(http.StatusNoContent)
}

func main() {
	// init database
	dataSourceName := "host=localhost user=postgres password=postgres port=5432 sslmode=disable"
	db, err = gorm.Open(postgres.Open(dataSourceName), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic(err)
	}

	// enable debug logging
	err = db.AutoMigrate(&CleaningTask{})
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
