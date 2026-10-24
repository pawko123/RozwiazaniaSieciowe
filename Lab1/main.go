package main

import (
	"Lab1/models"
	"Lab1/services"
	"flag"
	"fmt"
	"log"
	"sort"
	"strings"
)

func main() {
	teamFlag := flag.String("team", "", "Nazwa drużyny (lub fragment nazwy)")
	modeFlag := flag.String("mode", "all", "Rodzaj spotkań: home, away, all")
	flag.Parse()

	service := services.NewMatchRetrievalService("https://plk.pl/terminarz")

	fmt.Println("Pobieranie terminarza...")
	teams, err := service.FetchSchedule()
	if err != nil {
		log.Fatalf("Błąd serwisu: %v", err)
	}

	if *teamFlag == "" {
		fmt.Println("\nNie podano parametru -team. Dostępne drużyny:")
		var names []string
		for name := range teams {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			fmt.Printf(" - %s\n", name)
		}
		fmt.Println("\nUżycie:")
		fmt.Println("  go run . -team \"Legia\" -mode home")
		fmt.Println("  go run . -team \"Anwil\" -mode away")
		fmt.Println("  go run . -team \"Arka\" -mode all")
		return
	}

	var matchedTeam *models.Team
	if exactTeam, exists := teams[*teamFlag]; exists {
		matchedTeam = exactTeam
	} else {
		var candidates []*models.Team
		query := strings.ToLower(*teamFlag)

		for name, team := range teams {
			if strings.Contains(strings.ToLower(name), query) {
				candidates = append(candidates, team)
			}
		}

		switch len(candidates) {
		case 0:
			log.Fatalf("Nie znaleziono drużyny pasującej do: %q", *teamFlag)
		case 1:
			matchedTeam = candidates[0]
		default:
			fmt.Printf("Wyszukiwanie %q jest niejednoznaczne. Znaleziono kilka drużyn:\n", *teamFlag)
			for _, c := range candidates {
				fmt.Printf(" - %s\n", c.Name)
			}
			log.Fatal("Podaj dokładniejszą nazwę drużyny.")
		}
	}

	fmt.Printf("\nWyniki dla: %s (tryb: %s)\n", matchedTeam.Name, *modeFlag)
	fmt.Println(strings.Repeat("-", 80))

	var output []*models.Match
	switch strings.ToLower(*modeFlag) {
	case "home":
		output = matchedTeam.HomeMatches
	case "away":
		output = matchedTeam.AwayMatches
	case "all":
		output = append(output, matchedTeam.HomeMatches...)
		output = append(output, matchedTeam.AwayMatches...)
		sort.Slice(output, func(i, j int) bool {
			return output[i].Date.Before(output[j].Date)
		})
	default:
		log.Fatalf("Nieznany tryb: %q. Użyj: home, away lub all", *modeFlag)
	}

	if len(output) == 0 {
		fmt.Println("Brak meczów dla wybranych kryteriów.")
		return
	}

	for _, match := range output {
		fmt.Println(match)
	}
}
