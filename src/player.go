package src

const ROLECOUNT = 3
const MINIMUMPLAYERCOUNT = 5

type Role int

const (
	LIBERAL Role = iota
	HITLER
	FASCIST
)

type Player struct {
	RoomOwner bool
	Room      *Room
	Number    uint
	Name      string
	Job       Role
	Votes     uint
	Dead      bool
	Voted     bool
	Chosen    bool
}

func (pl *Player) Die() {
	if pl.Dead == false {
		pl.Dead = true
	}
}

func (pl *Player) AssignChosen() {
	if pl.Dead == false {
		pl.Chosen = true
	}
}

func (pl *Player) CreateRoom(roomName string) *Room {
	return CreateRoom(roomName, pl)
}

func (pl *Player) StartGame() bool {
	if pl.RoomOwner == true && len(pl.Room.players) >= MINIMUMPLAYERCOUNT {
		pl.Room.StartGame()
		return true
	}
	return false
}

func (pl *Player) ResetRound() {
	pl.Votes = 0
	pl.Voted = false
}

func (pl *Player) SetVotes(score uint) {
	pl.Votes = score
}

func (pl *Player) CastVote(votedPlayerName string) {
	if pl.Voted == false && pl.Dead == false {
		votedPlayer := pl.Room.FindPlayer(votedPlayerName)
		if votedPlayer != nil {
			votedPlayer.IncrementVote()
			pl.Voted = true
		}
	}
}

func (pl *Player) IncrementVote() {
	pl.Votes += 1
}

func (pl *Player) IsEligibleToChat() bool {
	if pl.Room != nil {
		if pl.Room.playing {
			if pl.Dead == true {
				return false
			} else if pl.Room.stage == FASCISTANDREST {
				return true
			} else if pl.Job == LIBERAL && pl.Room.stage == LIBERALSTAGE {
				return true
			} else if pl.Job == HITLER && pl.Room.stage == HITLERSTAGE {
				return true
			} else {
				return false
			}
		}
	}
	return true
}

func (pl *Player) End() {
	pl.RoomOwner = false
	pl.Job = FASCIST
	pl.Dead = false
	pl.Chosen = false
	pl.Voted = false
	pl.Votes = 0
	pl.Room = nil
}
