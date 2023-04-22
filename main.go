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

type CleaningTaskRepo struct {
	Db *gorm.DB
}

type Option func(*CleaningTaskRepo)

func WithDatabase(db *gorm.DB) Option {
	// assign & setup db into repo
	return func(repo *CleaningTaskRepo) {
		repo.Db = db
	}
}

func NewCleaningTaskRepo(options ...Option) *CleaningTaskRepo {
	repo := &CleaningTaskRepo{}
	// iterate over all options and apply them to the Repo
	for _, option := range options {
		option(repo)
	}
	return repo
}

func getCleaningTasks(c echo.Context, repo *CleaningTaskRepo) error {
	cleaningTasks := []CleaningTask{{}}
	result := repo.Db.Model(&CleaningTask{}).Find(&cleaningTasks)
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, result.Error)
	}
	return c.JSON(http.StatusOK, cleaningTasks)
}

func createCleaningTask(c echo.Context, repo *CleaningTaskRepo) error {
	cleaningTask := CleaningTask{}
	err := json.NewDecoder(c.Request().Body).Decode(&cleaningTask)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	// create cleaning task in db
	result := repo.Db.Create(&cleaningTask)

	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, result.Error)
	}

	return c.JSON(http.StatusOK, cleaningTask)
}

func getCleaningTaskById(c echo.Context, repo *CleaningTaskRepo) error {
	id := c.Param("id")
	cleaningTask := CleaningTask{}
	result := repo.Db.Model(CleaningTask{}).First(&cleaningTask, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound)
		}

		return echo.NewHTTPError(http.StatusInternalServerError, result.Error)
	}

	return c.JSON(http.StatusOK, cleaningTask)
}

func updateCleaningTaskById(c echo.Context, repo *CleaningTaskRepo) error {
	cleaningTask := CleaningTask{}
	err := json.NewDecoder(c.Request().Body).Decode(&cleaningTask)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	id, _ := strconv.Atoi(c.Param("id"))
	result := repo.Db.First(&CleaningTask{}, id)
	if result.RowsAffected == 0 {
		return c.NoContent(http.StatusNotFound)
	}
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, result.Error)
	}

	// save updated cleaning task in db
	cleaningTask.ID = uint(id)
	// TODO: include id in request
	result = repo.Db.Save(&cleaningTask)
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, result.Error)
	}

	return c.JSON(http.StatusOK, cleaningTask)
}

func deleteCleaningTaskById(c echo.Context, repo *CleaningTaskRepo) error {
	id := c.Param("id")
	result := repo.Db.Find(&CleaningTask{}, id)
	if result.RowsAffected == 0 {
		return c.NoContent(http.StatusNotFound)
	}
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, result.Error)
	}

	result = repo.Db.Delete(&CleaningTask{}, id)
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

	// enable debug logging
	err = db.AutoMigrate(&CleaningTask{})
	if err != nil {
		panic(err)
	}

	e := echo.New()

	repo := NewCleaningTaskRepo(WithDatabase(db))

	cleaningTasksPath := "/api/cleaning-tasks"

	e.GET(cleaningTasksPath, func(c echo.Context) error {
		return getCleaningTasks(c, repo)
	})
	e.POST(cleaningTasksPath, func(c echo.Context) error {
		return createCleaningTask(c, repo)
	})
	e.GET(cleaningTasksPath+"/:id", func(c echo.Context) error {
		return getCleaningTaskById(c, repo)
	})
	e.PUT(cleaningTasksPath+"/:id", func(c echo.Context) error {
		return updateCleaningTaskById(c, repo)
	})
	e.DELETE(cleaningTasksPath+"/:id", func(c echo.Context) error {
		return deleteCleaningTaskById(c, repo)
	})

	e.Logger.Fatal(e.Start(":1323"))
}
