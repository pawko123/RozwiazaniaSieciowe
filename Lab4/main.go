package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"lab4/entities"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	databasePath := flag.String("database", "gorm.db", "ścieżka do pliku bazy SQLite")
	firstName := flag.String("first", "", "imię osoby (razem z -last): roczne wypłaty osoby")
	lastName := flag.String("last", "", "nazwisko osoby (razem z -first): roczne wypłaty osoby")
	department := flag.String("department", "", "nazwa działu: roczne wypłaty działu")
	tax := flag.Float64("tax", 0, "wysokość podatku w procentach (razem z -month): miesięczny podatek firmy")
	month := flag.Int("month", int(time.Now().Month()), "miesiąc (1-12) dla -tax")
	year := flag.Int("year", time.Now().Year(), "rok kalendarzowy")
	flag.Parse()

	providedFlags := map[string]bool{}
	flag.Visit(func(providedFlag *flag.Flag) { providedFlags[providedFlag.Name] = true })

	_, statError := os.Stat(*databasePath)

	database, err := gorm.Open(sqlite.Open(*databasePath), &gorm.Config{})
	if err != nil {
		log.Fatalf("nie można otworzyć bazy: %v", err)
	}
	if err := database.AutoMigrate(&entities.UserEntity{}, &entities.DepartmentEntity{}, &entities.SalaryEntity{}); err != nil {
		log.Fatalf("migracja nie powiodła się: %v", err)
	}
	if errors.Is(statError, os.ErrNotExist) {
		if err := entities.Seed(database); err != nil {
			os.Remove(*databasePath)
			log.Fatalf("Zapełnianie bazy nie powiodło się: %v", err)
		}
		fmt.Printf("Utworzono i zapełniono bazę %s danymi przykładowymi\n", *databasePath)
	}

	departments := entities.NewDepartmentRepository(database)
	users := entities.NewUserRepository(database)
	salaries := entities.NewSalaryRepository(database)

	if *firstName != "" && *lastName != "" {
		total, err := users.CalculateUserYearlySalary(*firstName, *lastName, *year)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Roczne wypłaty %s %s w %d: %.2f\n", *firstName, *lastName, *year, total)
	}

	if *department != "" {
		total, err := departments.CalculateDepartmentYearlySalary(*department, *year)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Roczne wypłaty działu %s w %d: %.2f\n", *department, *year, total)
	}

	if providedFlags["tax"] && *month >= 1 && *month <= 12 {
		total, err := salaries.CalculateSalaryTax(float32(*tax), time.Month(*month), *year)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Podatek %.2f%% od wynagrodzeń firmy za %02d/%d: %.2f\n", *tax, *month, *year, total)
	}
}
