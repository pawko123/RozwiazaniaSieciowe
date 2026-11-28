package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type RepositoryTestSuite struct {
	suite.Suite
	database    *gorm.DB
	transaction *gorm.DB

	departmentRepository *DepartmentRepository
	userRepository       *UserRepository
	salaryRepository     *SalaryRepository
}

func TestRepositories(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}

func (testSuite *RepositoryTestSuite) SetupSuite() {
	database, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	testSuite.Require().NoError(err)

	sqlDatabase, err := database.DB()
	testSuite.Require().NoError(err)
	sqlDatabase.SetMaxOpenConns(1)

	testSuite.Require().NoError(database.AutoMigrate(&UserEntity{}, &DepartmentEntity{}, &SalaryEntity{}))
	testSuite.Require().NoError(Seed(database))
	testSuite.database = database
}

func (testSuite *RepositoryTestSuite) TearDownSuite() {
	sqlDatabase, err := testSuite.database.DB()
	testSuite.Require().NoError(err)
	testSuite.Require().NoError(sqlDatabase.Close())
}

func (testSuite *RepositoryTestSuite) SetupTest() {
	testSuite.transaction = testSuite.database.Begin()
	testSuite.Require().NoError(testSuite.transaction.Error)

	testSuite.departmentRepository = NewDepartmentRepository(testSuite.transaction)
	testSuite.userRepository = NewUserRepository(testSuite.transaction)
	testSuite.salaryRepository = NewSalaryRepository(testSuite.transaction)
}

func (testSuite *RepositoryTestSuite) TearDownTest() {
	testSuite.Require().NoError(testSuite.transaction.Rollback().Error)
}

func (testSuite *RepositoryTestSuite) findUser(firstName, lastName string) UserEntity {
	var user UserEntity
	testSuite.Require().NoError(testSuite.transaction.
		Where("firstName = ? AND lastName = ?", firstName, lastName).
		First(&user).Error)
	return user
}

func (testSuite *RepositoryTestSuite) TestCreateAndRelations() {
	var jan UserEntity
	testSuite.Require().NoError(testSuite.transaction.
		Preload("Departments").Preload("Salaries").
		Where("firstName = ? AND lastName = ?", "Jan", "Kowalski").
		First(&jan).Error)
	testSuite.Len(jan.Departments, 2)
	testSuite.Len(jan.Salaries, 24)

	var itDepartment DepartmentEntity
	testSuite.Require().NoError(testSuite.transaction.
		Preload("Users").
		Where("departmentName = ?", "IT").
		First(&itDepartment).Error)
	testSuite.Len(itDepartment.Users, 2)
}

func (testSuite *RepositoryTestSuite) TestCalculateDepartmentYearlySalary() {
	testCases := []struct {
		name           string
		departmentName string
		year           int
		expectedTotal  float32
	}{
		{"IT 2025", "IT", 2025, 62500 + 75000},
		{"HR 2025", "HR", 2025, 62500 + 56250},
		{"Finance 2026", "Finance", 2026, 87500},
		{"rok bez wypłat", "IT", 2024, 0},
		{"nieistniejący dział", "XX", 2025, 0},
	}
	for _, testCase := range testCases {
		actualTotal, err := testSuite.departmentRepository.CalculateDepartmentYearlySalary(testCase.departmentName, testCase.year)
		testSuite.Require().NoError(err, testCase.name)
		testSuite.InDelta(testCase.expectedTotal, actualTotal, 0.01, testCase.name)
	}
}

func (testSuite *RepositoryTestSuite) TestCalculateUserYearlySalary() {
	testCases := []struct {
		name          string
		firstName     string
		lastName      string
		year          int
		expectedTotal float32
	}{
		{"Jan 2025", "Jan", "Kowalski", 2025, 62500},
		{"Anna 2025", "Anna", "Nowak", 2025, 75000},
		{"Piotr 2026", "Piotr", "Zieliński", 2026, 56250},
		{"rok bez wypłat", "Jan", "Kowalski", 2024, 0},
		{"złe imię", "Adam", "Kowalski", 2025, 0},
	}
	for _, testCase := range testCases {
		actualTotal, err := testSuite.userRepository.CalculateUserYearlySalary(testCase.firstName, testCase.lastName, testCase.year)
		testSuite.Require().NoError(err, testCase.name)
		testSuite.InDelta(testCase.expectedTotal, actualTotal, 0.01, testCase.name)
	}
}

func (testSuite *RepositoryTestSuite) TestCalculateSalaryTax() {
	testCases := []struct {
		name        string
		month       time.Month
		year        int
		expectedTax float32
	}{
		{"styczeń 2025", time.January, 2025, 4275},
		{"grudzień 2025 (z premiami)", time.December, 2025, 6412.5},
		{"miesiąc bez wypłat", time.March, 2024, 0},
	}
	for _, testCase := range testCases {
		actualTax, err := testSuite.salaryRepository.CalculateSalaryTax(19, testCase.month, testCase.year)
		testSuite.Require().NoError(err, testCase.name)
		testSuite.InDelta(testCase.expectedTax, actualTax, 0.01, testCase.name)
	}
}

func (testSuite *RepositoryTestSuite) TestUpdateSalary() {
	jan := testSuite.findUser("Jan", "Kowalski")
	salaryDate := time.Date(2025, time.February, 10, 0, 0, 0, 0, time.UTC)

	testSuite.Require().NoError(testSuite.transaction.Model(&SalaryEntity{}).
		Where("userId = ? AND dateOfSalary = ?", jan.UserId, salaryDate).
		Updates(map[string]any{"salary": 1500, "bonus": 250}).Error)

	var updatedSalary SalaryEntity
	testSuite.Require().NoError(testSuite.transaction.
		Where("userId = ? AND dateOfSalary = ?", jan.UserId, salaryDate).
		First(&updatedSalary).Error)
	testSuite.Equal(float32(1500), updatedSalary.Salary)
	testSuite.Equal(float32(250), updatedSalary.Bonus)
}

func (testSuite *RepositoryTestSuite) TestUpdateUserDepartments() {
	anna := testSuite.findUser("Anna", "Nowak")

	var hrDepartment DepartmentEntity
	testSuite.Require().NoError(testSuite.transaction.Where("departmentName = ?", "HR").First(&hrDepartment).Error)
	testSuite.Require().NoError(testSuite.transaction.Model(&anna).Association("Departments").Append(&hrDepartment))

	var updatedAnna UserEntity
	testSuite.Require().NoError(testSuite.transaction.
		Preload("Departments").
		First(&updatedAnna, anna.UserId).Error)

	var departmentNames []string
	for _, department := range updatedAnna.Departments {
		departmentNames = append(departmentNames, department.DepartmentName)
	}
	testSuite.ElementsMatch([]string{"IT", "HR"}, departmentNames)
}
