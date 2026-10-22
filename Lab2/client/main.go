package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"Lab2/shared/protocol"
)

func main() {
	serverAddr := fmt.Sprintf("127.0.0.1:%d", protocol.DefaultPort)
	keyClient := NewClient(serverAddr)
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("==================================================")
	fmt.Println("             KLIENT BAZY KLUCZY (KEY)             ")
	fmt.Println("==================================================")
	fmt.Println("Wybierz operację:")
	fmt.Println("  1) Sprawdź obecność klucza")
	fmt.Println("  2) Dodaj nowy klucz")
	fmt.Print("Wybór [1-2]: ")

	if !scanner.Scan() {
		return
	}
	choice := strings.TrimSpace(scanner.Text())

	switch choice {
	case "1":
		fmt.Print("Podaj numer klucza: ")
		if !scanner.Scan() {
			return
		}
		keyID := strings.TrimSpace(scanner.Text())
		if keyID == "" {
			fmt.Println("[BŁĄD] Numer klucza nie może być pusty.")
			return
		}

		exists, err := keyClient.CheckKey(keyID)
		if err != nil {
			fmt.Printf("[BŁĄD] Nie powiodło się sprawdzenie klucza: %v\n", err)
		} else if exists {
			fmt.Printf("[WYNIK] Klucz '%s' ZNAJDUJE SIĘ w bazie.\n", keyID)
		} else {
			fmt.Printf("[WYNIK] Klucz '%s' NIE ZNAJDUJE SIĘ w bazie.\n", keyID)
		}

	case "2":
		fmt.Print("Podaj hasło administratora: ")
		if !scanner.Scan() {
			return
		}
		password := strings.TrimSpace(scanner.Text())

		fmt.Print("Podaj numer klucza do dodania: ")
		if !scanner.Scan() {
			return
		}
		keyID := strings.TrimSpace(scanner.Text())
		if keyID == "" {
			fmt.Println("[BŁĄD] Numer klucza nie może być pusty.")
			return
		}

		added, err := keyClient.AddKey(password, keyID)
		if err != nil {
			fmt.Printf("[BŁĄD] Nie powiodło się dodanie klucza: %v\n", err)
		} else if added {
			fmt.Printf("[WYNIK] Klucz '%s' został POMYŚLNIE DODANY do bazy.\n", keyID)
		} else {
			fmt.Println("[WYNIK] BŁĄD AUTORYZACJI: Operacja odrzucona przez serwer (błędne hasło).")
		}

	default:
		fmt.Printf("[BŁĄD] Nieznana opcja: '%s'. Dozwolone wartości to 1 lub 2.\n", choice)
	}

	fmt.Println("--------------------------------------------------")
	fmt.Println("Zakończono działanie.")
}
