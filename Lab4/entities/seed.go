package entities

import (
	"time"

	"gorm.io/gorm"
)

func Seed(database *gorm.DB) error {
	return database.Transaction(func(transaction *gorm.DB) error {
		departments := []*DepartmentEntity{
			{Address: "Zielona Góra, ul. Prosta 1", DepartmentName: "IT", Mail: "it@firma.pl", Phone: "111111111", Description: "Dział informatyki"},
			{Address: "Zielona Góra, ul. Prosta 2", DepartmentName: "HR", Mail: "hr@firma.pl", Phone: "222222222", Description: "Dział kadr"},
			{Address: "Zielona Góra, ul. Prosta 3", DepartmentName: "Finance", Mail: "finance@firma.pl", Phone: "333333333", Description: "Dział księgowy"},
		}
		if err := transaction.Create(departments).Error; err != nil {
			return err
		}
		itDepartment, hrDepartment, financeDepartment := *departments[0], *departments[1], *departments[2]

		type seedEmployee struct {
			user          UserEntity
			monthlySalary float32
		}
		employees := []seedEmployee{
			{UserEntity{FirstName: "Jan", LastName: "Kowalski", Username: "jkowalski", Password: "secret", Departments: []DepartmentEntity{itDepartment, hrDepartment}}, 5000},
			{UserEntity{FirstName: "Anna", LastName: "Nowak", Username: "anowak", Password: "secret", Departments: []DepartmentEntity{itDepartment}}, 6000},
			{UserEntity{FirstName: "Piotr", LastName: "Zieliński", Username: "pzielinski", Password: "secret", Departments: []DepartmentEntity{hrDepartment}}, 4500},
			{UserEntity{FirstName: "Maria", LastName: "Wiśniewska", Username: "mwisniewska", Password: "secret", Departments: []DepartmentEntity{financeDepartment}}, 7000},
		}

		for index := range employees {
			employee := &employees[index]
			if err := transaction.Create(&employee.user).Error; err != nil {
				return err
			}
			var salaries []SalaryEntity
			for _, year := range []int{2025, 2026} {
				for month := time.January; month <= time.December; month++ {
					var bonus float32
					if month == time.December {
						bonus = employee.monthlySalary / 2
					}
					salaries = append(salaries, SalaryEntity{
						UserId:       employee.user.UserId,
						Salary:       employee.monthlySalary,
						Bonus:        bonus,
						DateOfSalary: time.Date(year, month, 10, 0, 0, 0, 0, time.UTC),
					})
				}
			}
			if err := transaction.Create(&salaries).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
