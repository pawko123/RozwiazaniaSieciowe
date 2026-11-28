package entities

import (
	"time"

	"gorm.io/gorm"
)

type DepartmentEntity struct {
	DepartmentId   uint          `gorm:"primaryKey;autoIncrement;column:departmentId;"`
	Address        string        `gorm:"size:100;not null;column:address;"`
	DepartmentName string        `gorm:"size:100;not null;unique;column:departmentName;"`
	Mail           string        `gorm:"size:100;not null;unique;column:mail;"`
	Phone          string        `gorm:"size:10;not null;unique;column:phone;"`
	Description    string        `gorm:"size:255;column:description;"`
	Users          []*UserEntity `gorm:"many2many:UserDepartments;joinForeignKey:DepartmentId;joinReferences:UserId;"`
}

func (DepartmentEntity) TableName() string {
	return "DepartmentEntity"
}

type DepartmentRepository struct {
	database *gorm.DB
}

func NewDepartmentRepository(database *gorm.DB) *DepartmentRepository {
	return &DepartmentRepository{database: database}
}

func (repository *DepartmentRepository) CalculateDepartmentYearlySalary(departmentName string, year int) (float32, error) {
	yearStart := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	yearEnd := yearStart.AddDate(1, 0, 0)

	var total float32
	err := repository.database.Model(&SalaryEntity{}).
		Joins("JOIN user_departments ON user_departments.user_id = SalaryEntity.userId").
		Joins("JOIN DepartmentEntity ON DepartmentEntity.departmentId = user_departments.department_id").
		Where("DepartmentEntity.departmentName = ?", departmentName).
		Where("SalaryEntity.dateOfSalary >= ? AND SalaryEntity.dateOfSalary < ?", yearStart, yearEnd).
		Select("COALESCE(SUM(SalaryEntity.salary + SalaryEntity.bonus), 0)").
		Scan(&total).Error
	return total, err
}
