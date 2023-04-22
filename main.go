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

type CleaningTaskService struct {
	Db *gorm.DB
}

func (s *CleaningTaskService) getCleaningTasks(c echo.Context) error {
	cleaningTasks := []CleaningTask{{}}
	result := s.Db.Model(&CleaningTask{}).Find(&cleaningTasks)
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, result.Error)
	}
	return c.JSON(http.StatusOK, cleaningTasks)
}

func (s *CleaningTaskService) createCleaningTask(c echo.Context) error {
	cleaningTask := CleaningTask{}
	err := json.NewDecoder(c.Request().Body).Decode(&cleaningTask)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	// create cleaning task in db
	result := s.Db.Create(&cleaningTask)

	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, result.Error)
	}

	return c.JSON(http.StatusOK, cleaningTask)
}

func (s *CleaningTaskService) getCleaningTaskById(c echo.Context) error {
	id := c.Param("id")
	cleaningTask := CleaningTask{}
	result := s.Db.Model(CleaningTask{}).First(&cleaningTask, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound)
		}

		return echo.NewHTTPError(http.StatusInternalServerError, result.Error)
	}

	return c.JSON(http.StatusOK, cleaningTask)
}

func (s *CleaningTaskService) updateCleaningTaskById(c echo.Context) error {
	cleaningTask := CleaningTask{}
	err := json.NewDecoder(c.Request().Body).Decode(&cleaningTask)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	id, _ := strconv.Atoi(c.Param("id"))
	result := s.Db.First(&CleaningTask{}, id)
	if result.RowsAffected == 0 {
		return c.NoContent(http.StatusNotFound)
	}
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, result.Error)
	}

	// save updated cleaning task in db
	cleaningTask.ID = uint(id)
	// TODO: include id in request
	result = s.Db.Save(&cleaningTask)
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, result.Error)
	}

	return c.JSON(http.StatusOK, cleaningTask)
}

func (s *CleaningTaskService) deleteCleaningTaskById(c echo.Context) error {
	id := c.Param("id")
	result := s.Db.Find(&CleaningTask{}, id)
	if result.RowsAffected == 0 {
		return c.NoContent(http.StatusNotFound)
	}
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, result.Error)
	}

	result = s.Db.Delete(&CleaningTask{}, id)
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, result.Error)
	}

	return c.NoContent(http.StatusNoContent)
}

func main() {
	// init database
	dataSourceName := "host=localhost user=postgres password=postgres port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dataSourceName), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic(err)
	}

	// enable auto migrations
	err = db.AutoMigrate(&CleaningTask{})
	if err != nil {
		panic(err)
	}

	// setup CleaningTaskService with shared database dependency
	service := &CleaningTaskService{Db: db}

	e := echo.New()

	cleaningTasksPath := "/api/cleaning-tasks"

	e.GET(cleaningTasksPath, service.getCleaningTasks)
	e.POST(cleaningTasksPath, service.createCleaningTask)
	e.GET(cleaningTasksPath+"/:id", service.getCleaningTaskById)
	e.PUT(cleaningTasksPath+"/:id", service.updateCleaningTaskById)
	e.DELETE(cleaningTasksPath+"/:id", service.deleteCleaningTaskById)

	e.Logger.Fatal(e.Start(":1323"))
}
