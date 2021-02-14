package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"regexp"

	"github.com/nobodylovesm3/SecretHitler/src"
)

type Mesg struct {
	content string
	room    *src.Room
	special bool
}

func main() {
	ln, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		log.Println(err.Error())
	}
	allRooms := make([]*src.Room, 0)
	allClients := make(map[net.Conn]*src.Player)
	newConnections := make(chan net.Conn)
	deadConnections := make(chan net.Conn)
	messages := make(chan Mesg)

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Println(err.Error())
			}
			conn.Write([]byte("Input your username: "))
			nameByte := make([]byte, 1024)
			go func() {
				readBytes, err := conn.Read(nameByte)
				if err != nil {
					log.Println(err.Error())
				} else {
					name := string(nameByte[:readBytes-1])
					for !isValidName(name, allClients) {
						conn.Write([]byte(name + " is taken, please choose another name!"))
						readBytes, err := conn.Read(nameByte)
						if err != nil {
							log.Println(err.Error())
						}
						name = string(nameByte[:readBytes-1])
					}
					allClients[conn] = &src.Player{false, nil, src.ClientCount,
						name, src.FASCIST,
						0, false, false, false}
					newConnections <- conn
					messages <- Mesg{fmt.Sprintln(allClients[conn].Name, " joined the room!"),
						nil, false}
				}
			}()
		}
	}()

	for {
		select {
		case curCon := <-newConnections:
			src.ClientCount += 1
			go func(conn net.Conn) {
				rd := bufio.NewReader(conn)
				for {
					curPlayer := allClients[conn]
					curRoom := allClients[conn].Room
					if !(curPlayer.IsEligibleToChat()) {
						continue
					}
					mesg, err := rd.ReadString('\n')
					if err != nil {
						break
					}
					if curRoom == nil {
						joinRoomReg := regexp.MustCompile(`&join-room (\w+)`)
						createRoomReg := regexp.MustCompile(`&create-room (?P<RoomName>\w+)`)
						allRoomsReg := regexp.MustCompile(`&show-rooms`)
						if createRoomReg.MatchString(mesg) == true {
							res := createRoomReg.FindStringSubmatch(mesg)
							newRoom := &src.Room{}
							newRoom.SetName(res[1])
							newRoom.AddPlayer(curPlayer)
							allRooms = append(allRooms, newRoom)
							conn.Write([]byte("You have successfully created room " + res[1]))
							continue
						} else if joinRoomReg.MatchString(mesg) == true {
							res := joinRoomReg.FindStringSubmatch(mesg)
							exists := false
							for _, r := range allRooms {
								if r.GetName() == res[1] && !(r.IsPlaying()) {
									exists = true
									r.AddPlayer(curPlayer)
									conn.Write([]byte("You have successfully joined room " + r.GetName()))
									messages <- Mesg{fmt.Sprintln("\n", curPlayer.Name,
										" joined ", r.GetName()),
										r, false}
								}
							}
							if exists == false {
								conn.Write([]byte("Room is non-existant!"))
							}
							continue
						} else if allRoomsReg.MatchString(mesg) == true {
							var buffer bytes.Buffer
							var srco string
							for _, r := range allRooms {
								if r != nil {
									srco = r.GetName() + " owned by " + r.GetOwner().Name + "\n"
									if r.IsPlaying() {
										srco = "Playing: " + srco
									} else {
										srco = "Lobby: " + srco
									}
									buffer.WriteString(srco)
								}
							}
							if buffer.Len() == 0 {
								conn.Write([]byte("There are no rooms available!\n"))
							} else {
								conn.Write([]byte("The currently available rooms are: \n" + buffer.String()))
							}
							continue
						}
					} else if curRoom != nil && !(curRoom.IsPlaying()) && curPlayer.RoomOwner {
						startGameReg := regexp.MustCompile(`%start`)
						if startGameReg.MatchString(mesg) == true {
							outcome := curPlayer.StartGame()
							if outcome {
								messages <- Mesg{"", curRoom, true}
								messages <- Mesg{"The game has begun!\n FASCISTS - VOTE NOW!\n", curRoom, true}
							} else {
								conn.Write([]byte("Room can't be started!\n"))
							}
							continue
						}
					} else if curRoom.IsPlaying() {

						voteReg := regexp.MustCompile(`%choose (\w+)`)
						getPlayersReg := regexp.MustCompile(`%choose`)
						if voteReg.MatchString(mesg) == true {
							matchRes := voteReg.FindStringSubmatch(mesg)
							votedPlayerName := matchRes[1]
							curPlayer.CastVote(votedPlayerName)
							curStage := curRoom.GetStage()

							if curRoom.CanGoToNextStage() {
								hotSeatPlayer := curRoom.GetMostVotedPlayer()
								switch curStage {
								case src.LIBERALSTAGE:
									hotSeatPlayer.AssignChosen()
								case src.HITLERSTAGE:
									if curRoom.HasHitler() {
										//must add laws
									}
								case src.FASCISTANDREST:
									hotSeatPlayer.Die()
									messages <- Mesg{"Go town has decided to imprison " +
										hotSeatPlayer.Name + "\n", curRoom, false}
								}
								curRoom.NextStage()
								if res, winner := curRoom.GameOver(); res {
									if winner == src.LIBERAL {
										messages <- Mesg{"The LIBERALS HAVE WON!\n", curRoom, true}
									} else if winner == src.FASCIST {
										messages <- Mesg{"The FASCISTS HAVE WON!\n", curRoom, true}
									}
									indexOfCurRoom := findIndex(curRoom, allRooms)
									curRoom.End()
									allRooms = append(allRooms[:indexOfCurRoom], allRooms[indexOfCurRoom+1:]...)
									continue
								}
								curRoom.Reset()
								if curRoom.GetStage() == src.LIBERALSTAGE {
									messages <- Mesg{"LIBERALS!\n", curRoom, true}
									messages <- Mesg{"DECIDE WHO MIGHT BE HITLER!\n",
										curRoom, false}
									continue
								} else if curRoom.GetStage() == src.HITLERSTAGE {
									if curRoom.HasHitler() {
										messages <- Mesg{"HITLER TIME!", curRoom, true}
										messages <- Mesg{"CHOOSE A LAW TO APPLY!\n",
											curRoom, false}
									} else {
										curRoom.NextStage()
										curRoom.Reset()
										deadGuy := curRoom.FindChosenPlayerToDie()
										var announcement2 string
										if deadGuy != nil {
											announcement2 = deadGuy.Name +
												" has been killed!\n"
										}
										messages <- Mesg{announcement2,
											curRoom, false}
									}
									continue
								} else {
									deadPlayer := curRoom.FindChosenPlayerToDie()
									var announcement string
									if deadPlayer != nil {
										announcement = deadPlayer.Name +
											" has been killed!\n"
									} else {
										announcement = "Special law! Not implemented :( !\n"
									}
									messages <- Mesg{announcement,
										curRoom, false}
									continue
								}
							}
						} else if getPlayersReg.MatchString(mesg) == true {
							var playersBuffer bytes.Buffer
							for _, pl := range curRoom.GetPlayers() {
								if pl.Dead == false {
									playersBuffer.WriteString(pl.Name + "\n")
								}
							}
							if playersBuffer.Len() == 0 {
								conn.Write([]byte("There are no players available!\n"))
							} else {
								conn.Write([]byte("The current players are: \n" + playersBuffer.String()))
							}
							continue
						}
					}
					messages <- Mesg{fmt.Sprintln("\n", curPlayer.Name, " : ", mesg),
						curRoom, false}
				}
				deadConnections <- conn
			}(curCon)
		case msg := <-messages:
			for conn, client := range allClients {
				if msg.room != nil && msg.room == client.Room && !(msg.room.IsPlaying()) {
					conn.Write([]byte(msg.content))
				} else if msg.room == nil && client.Room == nil {
					conn.Write([]byte(msg.content))
				} else if msg.room != nil && msg.room == client.Room && msg.room.IsPlaying() {
					currentStage := msg.room.GetStage()
					if msg.special {
						if len(msg.content) == 0 {
							curJob := client.Job
							var roleDescription string
							switch curJob {
							case src.LIBERAL:
								roleDescription = "You are a LIBERAL!\n"
							case src.FASCIST:
								roleDescription = "You are a FASCIST!\n"
							case src.HITLER:
								roleDescription = "You are the HITLER!\n"
							}
							conn.Write([]byte(roleDescription))
						} else {
							conn.Write([]byte(msg.content))
						}
					} else if client.Dead {
						conn.Write([]byte("Dead Chat -> " + msg.content))
					} else if currentStage == src.LIBERALSTAGE {
						if client.Job == src.LIBERAL {
							conn.Write([]byte(msg.content))
						}
					} else if currentStage == src.HITLERSTAGE {
						if client.Job == src.HITLER {
							conn.Write([]byte(msg.content))
						}
					} else if currentStage == src.FASCISTANDREST {
						conn.Write([]byte(msg.content))
					}
				}
			}
		case lostClient := <-deadConnections:
			gonePlayer := allClients[lostClient]
			go func(playerName string) {
				if gonePlayer.Room != nil {
					gonePlayer.Room.KickPlayer(gonePlayer)
				}
				messages <- Mesg{fmt.Sprintln("\n", playerName, " left"),
					gonePlayer.Room, true}
			}(gonePlayer.Name)
			delete(allClients, lostClient)
		}
	}
}

func findIndex(room *src.Room, rooms []*src.Room) int {
	for index, v := range rooms {
		if v == room {
			return index
		}
	}
	return -1
}

func isValidName(s string, players map[net.Conn]*src.Player) bool {
	for _, pl := range players {
		if s == pl.Name {
			return false
		}
	}
	return true
}
