package controller

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
	cleaningTaskRepository *database.CleaningTaskRepository
)

func SetupCleaningTaskEndpoints(e *echo.Echo, repository *database.CleaningTaskRepository) {
	cleaningTaskRepository = repository

	cleaningTasksPath := "/api/cleaning-tasks"

	e.GET(cleaningTasksPath, func(c echo.Context) error {
		cleaningTasks, err := cleaningTaskRepository.FindAll()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}
		return c.JSON(http.StatusOK, cleaningTasks)
	})

	e.POST(cleaningTasksPath, func(c echo.Context) error {
		cleaningTask := models.CleaningTask{}
		err := json.NewDecoder(c.Request().Body).Decode(&cleaningTask)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}

		// create cleaning task in db
		err = cleaningTaskRepository.Save(&cleaningTask)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}

		return c.JSON(http.StatusOK, cleaningTask)
	})

	e.GET(cleaningTasksPath+"/:id", func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Unable to parse id!")
		}

		cleaningTask, err := cleaningTaskRepository.FindById(uint(id))
		if err == nil {
			return c.JSON(http.StatusOK, cleaningTask)
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound)
		}

		return echo.NewHTTPError(http.StatusInternalServerError, err)
	})

	e.PUT(cleaningTasksPath+"/:id", func(c echo.Context) error {
		cleaningTask := models.CleaningTask{}
		err := json.NewDecoder(c.Request().Body).Decode(&cleaningTask)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Unable to parse id!")
		}

		_, err = cleaningTaskRepository.FindById(uint(id))
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, err)
		}

		// TODO: include id in request
		cleaningTask.ID = uint(id)
		err = cleaningTaskRepository.Save(&cleaningTask)

		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}

		return c.JSON(http.StatusOK, cleaningTask)
	})

	e.DELETE(cleaningTasksPath+"/:id", func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Unable to parse id!")
		}

		_, err = cleaningTaskRepository.FindById(uint(id))
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, err)
		}

		err = cleaningTaskRepository.Delete(uint(id))
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}

		return c.NoContent(http.StatusNoContent)
	})
}
