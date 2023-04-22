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

func NewCleaningTaskRepo(db *gorm.DB, options ...Option) *CleaningTaskRepo {
	repo := &CleaningTaskRepo{
		Db: db,
	}
	// iterate over all options and apply them to the Repo
	for _, option := range options {
		option(repo)
	}
	return repo
}

func (repo *CleaningTaskRepo) GetAll() ([]CleaningTask, error) {
	cleaningTasks := []CleaningTask{{}}
	result := repo.Db.Model(&CleaningTask{}).Find(&cleaningTasks)
	if result.Error != nil {
		return nil, result.Error
	}
	return cleaningTasks, nil
}

func (repo *CleaningTaskRepo) Create(task *CleaningTask) error {
	return repo.Db.Create(task).Error
}

func (repo *CleaningTaskRepo) GetById(id uint) (*CleaningTask, error) {
	task := &CleaningTask{}
	result := repo.Db.Model(CleaningTask{}).First(task, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return task, nil
}

func (repo *CleaningTaskRepo) Update(task *CleaningTask) error {
	return repo.Db.Save(task).Error
}

func (repo *CleaningTaskRepo) DeleteById(id uint) error {
	return repo.Db.Delete(&CleaningTask{}, id).Error
}

// echo handlers

func getCleaningTasks(c echo.Context, repo *CleaningTaskRepo) error {
	cleaningTasks, err := repo.GetAll()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, cleaningTasks)
}

func createCleaningTask(c echo.Context, repo *CleaningTaskRepo) error {
	cleaningTask := CleaningTask{}
	err := json.NewDecoder(c.Request().Body).Decode(&cleaningTask)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	err = repo.Create(&cleaningTask)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, cleaningTask)
}

func getCleaningTaskById(c echo.Context, repo *CleaningTaskRepo) error {
	id, _ := strconv.Atoi(c.Param("id"))
	cleaningTask, err := repo.GetById(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err)
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
	cleaningTask.ID = uint(id)

	err = repo.Update(&cleaningTask)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, cleaningTask)
}

func deleteCleaningTaskById(c echo.Context, repo *CleaningTaskRepo) error {
	id, _ := strconv.Atoi(c.Param("id"))

	err := repo.DeleteById(uint(id))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
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

	repo := NewCleaningTaskRepo(db)

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
