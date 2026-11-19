package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"Lab3/peer"
	"Lab3/ui"
)

func main() {
	nicknameFlag := flag.String("nick", "", "Pseudonim uzytkownika (wymagany)")
	multicastAddressFlag := flag.String("addr", "239.255.43.21:9999", "Adres grupy multicastowej oraz port")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Uzycie: go run . -nick <pseudonim> [opcje]")
		fmt.Fprintln(os.Stderr, "Dostepne flagi:")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *nicknameFlag == "" {
		fmt.Fprintln(os.Stderr, "Blad: Wymagane jest podanie pseudonimu za pomoca flagi -nick.")
		fmt.Fprintln(os.Stderr)
		flag.Usage()
		os.Exit(1)
	}

	var configuration = peer.NodeConfiguration{
		Nickname:           *nicknameFlag,
		MulticastAddress:   *multicastAddressFlag,
		HealthcheckPeriod:  2 * time.Second,
		UserTimeoutPeriod:  6 * time.Second,
		ProbeTimeoutPeriod: 350 * time.Millisecond,
		DedupCacheTTL:      60 * time.Second,
	}

	networkNode, nodeCreationError := peer.NewNode(configuration)
	if nodeCreationError != nil {
		fmt.Fprintf(os.Stderr, "Blad inicjalizacji wezla sieciowego: %v\n", nodeCreationError)
		os.Exit(1)
	}
	defer networkNode.LeaveAndClose()

	probeError := networkNode.ProbeAndJoin()
	if probeError != nil {
		if errors.Is(probeError, peer.ErrorNicknameTaken) {
			fmt.Fprintf(os.Stderr, "Blad: Pseudonim '%s' jest juz zajety w tej grupie multicastowej. Wybierz inny pseudonim.\n", *nicknameFlag)
		} else {
			fmt.Fprintf(os.Stderr, "Blad podczas dolaczania do grupy: %v\n", probeError)
		}
		os.Exit(1)
	}

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChannel
		networkNode.LeaveAndClose()
		os.Exit(0)
	}()

	userInterface := ui.NewUserInterface(networkNode, *nicknameFlag)
	runningError := userInterface.Run()
	if runningError != nil {
		fmt.Fprintf(os.Stderr, "Blad dzialania interfejsu uzytkownika: %v\n", runningError)
		os.Exit(1)
	}
}
