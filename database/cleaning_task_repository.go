package database

import (
	"github.com/timotheus90/go-playground/models"
)

type CleaningTaskRepository struct {
	DB *Database
}

func (repo *CleaningTaskRepository) Save(cleaningTask *models.CleaningTask) error {
	return repo.DB.Save(cleaningTask).Error
}

func (repo *CleaningTaskRepository) FindAll() ([]models.CleaningTask, error) {
	cleaningTasks := []models.CleaningTask{{}}
	result := repo.DB.Model(&models.CleaningTask{}).Find(&cleaningTasks)
	return cleaningTasks, result.Error
}

func (repo *CleaningTaskRepository) FindById(id uint) (*models.CleaningTask, error) {
	cleaningTask := models.CleaningTask{}
	result := repo.DB.Model(models.CleaningTask{}).First(&cleaningTask, id)
	return &cleaningTask, result.Error
}

func (repo *CleaningTaskRepository) Delete(id uint) error {
	return repo.DB.Delete(&models.CleaningTask{}, id).Error
}
