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

// DependencyContainer holds all the dependencies for the application.
type DependencyContainer struct {
	Db   *gorm.DB
	Echo *echo.Echo
}

func NewDependencyContainer() (*DependencyContainer, error) {
	dataSourceName := "host=localhost user=postgres password=postgres port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dataSourceName), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&CleaningTask{})
	if err != nil {
		return nil, err
	}

	e := echo.New()

	return &DependencyContainer{Db: db, Echo: e}, nil
}

func getTasksHandler(c echo.Context, dc *DependencyContainer) error {
	cleaningTasks := []CleaningTask{{}}
	result := dc.Db.Model(&CleaningTask{}).Find(&cleaningTasks)
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, result.Error)
	}
	return c.JSON(http.StatusOK, cleaningTasks)
}

func createTaskHandler(c echo.Context, dc *DependencyContainer) error {
	cleaningTask := CleaningTask{}
	err := json.NewDecoder(c.Request().Body).Decode(&cleaningTask)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	// create cleaning task in db
	result := dc.Db.Create(&cleaningTask)

	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, result.Error)
	}

	return c.JSON(http.StatusOK, cleaningTask)
}

func getTaskByIDHandler(c echo.Context, dc *DependencyContainer) error {
	id := c.Param("id")
	cleaningTask := CleaningTask{}
	result := dc.Db.Model(CleaningTask{}).First(&cleaningTask, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound)
		}

		return echo.NewHTTPError(http.StatusInternalServerError, result.Error)
	}

	return c.JSON(http.StatusOK, cleaningTask)
}

func updateTaskHandler(c echo.Context, dc *DependencyContainer) error {
	cleaningTask := CleaningTask{}
	err := json.NewDecoder(c.Request().Body).Decode(&cleaningTask)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	id, _ := strconv.Atoi(c.Param("id"))
	result := dc.Db.First(&CleaningTask{}, id)
	if result.RowsAffected == 0 {
		return c.NoContent(http.StatusNotFound)
	}
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, result.Error)
	}

	// save updated cleaning task in db
	cleaningTask.ID = uint(id)
	// TODO: include id in request
	result = dc.Db.Save(&cleaningTask)
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, result.Error)
	}

	return c.JSON(http.StatusOK, cleaningTask)
}

func deleteTaskHandler(c echo.Context, dc *DependencyContainer) error {
	id := c.Param("id")
	result := dc.Db.Find(&CleaningTask{}, id)
	if result.RowsAffected == 0 {
		return c.NoContent(http.StatusNotFound)
	}
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, result.Error)
	}

	result = dc.Db.Delete(&CleaningTask{}, id)
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, result.Error)
	}

	return c.NoContent(http.StatusNoContent)
}

func main() {
	dc, err := NewDependencyContainer()
	if err != nil {
		panic(err)
	}

	cleaningTasksPath := "/api/cleaning-tasks"

	// Pass the DependencyContainer to route handlers
	dc.Echo.GET(cleaningTasksPath, func(c echo.Context) error {
		return getTasksHandler(c, dc)
	})
	dc.Echo.POST(cleaningTasksPath, func(c echo.Context) error {
		return createTaskHandler(c, dc)
	})
	dc.Echo.GET(cleaningTasksPath+"/:id", func(c echo.Context) error {
		return getTaskByIDHandler(c, dc)
	})
	dc.Echo.PUT(cleaningTasksPath+"/:id", func(c echo.Context) error {
		return updateTaskHandler(c, dc)
	})
	dc.Echo.DELETE(cleaningTasksPath+"/:id", func(c echo.Context) error {
		return deleteTaskHandler(c, dc)
	})

	dc.Echo.Logger.Fatal(dc.Echo.Start(":1323"))
}
