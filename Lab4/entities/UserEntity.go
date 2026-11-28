package entities

import (
	"time"

	"gorm.io/gorm"
)

type UserEntity struct {
	UserId      uint               `gorm:"primaryKey;autoIncrement;column:userId;"`
	FirstName   string             `gorm:"size:100;not null;column:firstName;"`
	LastName    string             `gorm:"size:100;not null;column:lastName;"`
	Username    string             `gorm:"size:100;not null;unique;column:username;"`
	Password    string             `gorm:"size:60;not null;column:password;"`
	Description string             `gorm:"size:255;column:description;"`
	Salaries    []SalaryEntity     `gorm:"foreignKey:UserId;references:UserId;"`
	Departments []DepartmentEntity `gorm:"many2many:UserDepartments;joinForeignKey:UserId;joinReferences:DepartmentId;"`
}

func (UserEntity) TableName() string {
	return "UserEntity"
}

type UserRepository struct {
	database *gorm.DB
}

func NewUserRepository(database *gorm.DB) *UserRepository {
	return &UserRepository{database: database}
}

func (repository *UserRepository) CalculateUserYearlySalary(firstName string, lastName string, year int) (float32, error) {
	yearStart := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	yearEnd := yearStart.AddDate(1, 0, 0)

	var total float32
	err := repository.database.Model(&SalaryEntity{}).
		Joins("JOIN UserEntity ON UserEntity.userId = SalaryEntity.userId").
		Where("UserEntity.firstName = ? AND UserEntity.lastName = ?", firstName, lastName).
		Where("SalaryEntity.dateOfSalary >= ? AND SalaryEntity.dateOfSalary < ?", yearStart, yearEnd).
		Select("COALESCE(SUM(SalaryEntity.salary + SalaryEntity.bonus), 0)").
		Scan(&total).Error
	return total, err
}
