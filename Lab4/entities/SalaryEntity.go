package entities

import (
	"time"

	"gorm.io/gorm"
)

type SalaryEntity struct {
	SalaryId     uint      `gorm:"primaryKey;autoIncrement;column:salaryId;"`
	Bonus        float32   `gorm:"column:bonus;"`
	DateOfSalary time.Time `gorm:"type:date;column:dateOfSalary;"`
	Salary       float32   `gorm:"column:salary;"`
	UserId       uint      `gorm:"not null;index;column:userId;"`
}

func (SalaryEntity) TableName() string {
	return "SalaryEntity"
}

type SalaryRepository struct {
	database *gorm.DB
}

func NewSalaryRepository(database *gorm.DB) *SalaryRepository {
	return &SalaryRepository{database: database}
}

func (repository *SalaryRepository) CalculateSalaryTax(taxPercentage float32, month time.Month, year int) (float32, error) {
	monthStart := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, 0)

	var total float32
	err := repository.database.Model(&SalaryEntity{}).
		Where("dateOfSalary >= ? AND dateOfSalary < ?", monthStart, monthEnd).
		Select("COALESCE(SUM(salary + bonus), 0)").
		Scan(&total).Error
	if err != nil {
		return 0, err
	}
	return total * taxPercentage / 100, nil
}
