package models

type Team struct {
	Name        string
	HomeMatches []*Match
	AwayMatches []*Match
}

func NewTeam(name string) *Team {
	return &Team{
		Name:        name,
		HomeMatches: make([]*Match, 0),
		AwayMatches: make([]*Match, 0),
	}
}
