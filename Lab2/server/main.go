package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"Lab2/server/storage"
	"Lab2/shared/protocol"
)

func main() {
	port := flag.Int("port", protocol.DefaultPort, "Numer portu TCP do nasłuchiwania")
	pass := flag.String("pass", "secret123", "Hasło administratora do dodawania kluczy (set)")
	keysFlag := flag.String("keys", "123456789,admin_key", "Początkowa lista kluczy rozdzielona przecinkami")
	flag.Parse()

	store := storage.NewKeyStore()
	if *keysFlag != "" {
		for rawKey := range strings.SplitSeq(*keysFlag, ",") {
			if trimmedKey := strings.TrimSpace(rawKey); trimmedKey != "" {
				store.Add(trimmedKey)
			}
		}
	}
	handler := NewHandler(store, *pass)
	addr := fmt.Sprintf("0.0.0.0:%d", *port)
	srv := NewServer(addr, handler)

	fmt.Println("==================================================")
	fmt.Println("               SERWER PROTOKOŁU KEY               ")
	fmt.Println("==================================================")
	fmt.Printf(" Adres nasłuchu : %s\n", addr)
	fmt.Printf(" Hasło admina   : %s\n", *pass)
	fmt.Printf(" Baza kluczy    : %v\n", *keysFlag)
	fmt.Println("==================================================")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("[BŁĄD KRYTYCZNY] Serwer zakończył działanie z błędem: %v", err)
		}
	}()

	<-stop
	fmt.Println("\n[SERWER] Otrzymano sygnał zatrzymania, zamykanie serwera...")
	if err := srv.Close(); err != nil {
		log.Printf("[SERWER] Błąd podczas zamykania: %v", err)
	}
	fmt.Println("[SERWER] Serwer został pomyślnie wyłączony.")
}
