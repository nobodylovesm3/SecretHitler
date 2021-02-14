package src

import (
	"fmt"
	"math/rand"
	"time"
)

type Room struct {
	name    string
	players []*Player
	playing bool
	stage   int
}

const (
	LIBERALSTAGE int = iota
	HITLERSTAGE
	FASCISTANDREST
)

func (r Room) String() string {
	return fmt.Sprintf("The name of the room is %v and the players are %v", r.name, r.players)
}

func CreateRoom(name string, player *Player) *Room {
	resultRoom := &Room{name, nil, false, 0}
	newPlayers := make([]*Player, 0)
	newPlayers = append(newPlayers, player)
	resultRoom.players = newPlayers
	return resultRoom
}

func (r *Room) AddPlayer(player *Player) {
	if len(r.players) == 0 {
		player.RoomOwner = true
	}
	player.Room = r
	r.players = append(r.players, player)
}

var ClientCount uint = 0

func (r *Room) StartGame() {
	rand.Seed(time.Now().UnixNano())
	amountOfPlayers := len(r.players)
	doctorChoice := rand.Intn(amountOfPlayers)
	r.players[doctorChoice].Job = HITLER
	for index := range r.players {
		if index%2 == 0 && r.players[index].Job != HITLER {
			r.players[index].Job = LIBERAL
		} else if r.players[index].Job != HITLER {
			r.players[index].Job = FASCIST
		}
	}
	r.playing = true
	r.stage = 0
}

func (r *Room) IsPlaying() bool {
	return r.playing
}

func FindRoom(rooms *[]Room, name string) *Room {
	for _, rm := range *rooms {
		if rm.name == name {
			return &rm
		}
	}
	return nil
}

func (r *Room) FindPlayer(name string) *Player {
	for _, rm := range r.players {
		if rm.Name == name {
			return rm
		}
	}
	return nil
}

func (r *Room) Reset() {
	for _, player := range r.players {
		player.ResetRound()
	}
}

func (r *Room) GetOwner() *Player {
	for _, pl := range r.players {
		if pl.RoomOwner == true {
			return pl
		}
	}
	return nil
}

func (r *Room) GetPlayers() []*Player {
	return r.players
}

func (r *Room) SetName(name string) {
	r.name = name
}

func (r *Room) GetName() string {
	return r.name
}

func (r *Room) GetStage() int {
	return r.stage
}

func (r *Room) CanGoToNextStage() bool {
	if (r.stage == LIBERALSTAGE && r.CheckIfLiberalsVoted()) ||
		(r.stage == FASCISTANDREST && r.CheckIfAllVoted()) ||
		(r.stage == HITLERSTAGE && (!(r.HasHitler()))) {
		return true
	}
	return false
}

func (r *Room) NextStage() {
	if r.CanGoToNextStage() {
		r.stage++
		r.stage %= 3
	}
}

func (r *Room) GetMostVotedPlayer() *Player {
	maxVotedPlayer := r.players[0]
	for _, pl := range r.players {
		if pl.Votes > maxVotedPlayer.Votes {
			maxVotedPlayer = pl
		}
	}
	return maxVotedPlayer
}

func (r *Room) CheckIfAllVoted() bool {
	for _, pl := range r.players {
		if r.stage == FASCISTANDREST && pl.Voted == false && pl.Dead == false {
			return false
		}
	}
	return true
}

func (r *Room) CheckIfLiberalsVoted() bool {
	for _, pl := range r.players {
		if pl.Dead == false && pl.Job == LIBERAL && r.stage == LIBERALSTAGE && pl.Voted == false {
			return false
		}
	}
	return true
}

func (r *Room) FindChosenPlayerToDie() *Player {
	for index := range r.players {
		if r.players[index].Dead == false && r.players[index].Chosen == true {
			r.players[index].Dead = true
			return r.players[index]
		}
	}
	return nil
}

func (r *Room) HasHitler() bool {
	for _, pl := range r.players {
		if pl.Job == HITLER && pl.Dead == false {
			return true
		}
	}
	return false
}

func (r *Room) GameOver() (bool, Role) {
	aliveLiberals := 0
	aliveFascists := 0
	for _, pl := range r.players {
		if pl.Job == LIBERAL && pl.Dead == false {
			aliveLiberals++
		} else if (pl.Job == FASCIST || pl.Job == HITLER) && pl.Dead == false {
			aliveFascists++
		}
	}
	if aliveLiberals == 0 {
		return true, FASCIST
	} else if aliveFascists <= 1 && aliveLiberals >= 1 {
		return true, LIBERAL
	} else {
		return false, 0
	}
}

func (r *Room) End() {
	for index := range r.players {
		r.players[index].End()
	}
	r.playing = false
	r = nil
}

func (r *Room) KickPlayer(pl *Player) {
	wasRoomOwner := pl.RoomOwner
	for index := range r.players {
		if r.players[index] == pl {
			r.players[index].Dead = true
			if wasRoomOwner {
				r.players = append(r.players[:index], r.players[(index+1):]...)
				if len(r.players) != 0 {
					r.players[index+1].RoomOwner = true
					return
				}
			} else {
				r.players = append(r.players[:index], r.players[(index+1):]...)
				return
			}
		}
	}
}
