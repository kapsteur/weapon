package main

import (
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/kapsteur/messenger"
	"github.com/vrischmann/envconfig"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

type Config struct {
	VerifyToken string `envconfig:"VERIFY_TOKEN"`
	Verify      bool   `envconfig:"VERIFY,optional"`
	PageToken   string `envconfig:"PAGE_TOKEN"`
}

type Game struct {
	UserId            int64
	UserWin           int
	BotWin            int
	LastUserWeaponIdx int
	LastBotWeaponIdx  int
	IsWon             bool
}

var (
	conf    *Config
	weapons = []string{"✌", "✊", "✋"}
	replies = []messenger.QuickReply{
		messenger.QuickReply{ContentType: "text", Title: "✌", Payload: "1"},
		messenger.QuickReply{ContentType: "text", Title: "✊", Payload: "2"},
		messenger.QuickReply{ContentType: "text", Title: "✋", Payload: "3"}}
	rules = [][]int{[]int{0, -1, 1}, []int{1, 0, -1}, []int{-1, 1, 0}}
)

func init() {

	envconfig.Init(&conf)

	// Create a new messenger client
	client := messenger.New(messenger.Options{
		Verify:      conf.Verify,
		VerifyToken: conf.VerifyToken,
		Token:       conf.PageToken,
	})

	// Setup a handler to be triggered when a message is received
	client.HandleMessage(func(req *http.Request, m messenger.Message, r *messenger.Response) {
		ctx := appengine.NewContext(req)
		rand.Seed(time.Now().UnixNano())

		log.Infof(ctx, "%d: %v (Sent, %v)", m.Sender.ID, m.Text, m.Time.Format(time.UnixDate))

		game := Game{UserId: m.Sender.ID}
		gameKey := datastore.NewKey(ctx, "Game", "", m.Sender.ID, nil)
		err := datastore.Get(ctx, gameKey, &game)
		if err != nil {
			log.Infof(ctx, "Get - Err:%s", err)
		}

		p, err := client.ProfileByID(ctx, m.Sender.ID)
		if err != nil {
			log.Infof(ctx, "ProfileByID - Err:%s", err)
		}

		if m.Text == "✌" || m.Text == "✊" || m.Text == "✋" {

			userWeaponIdx := 0
			for idx, w := range weapons {
				if w == m.Text {
					userWeaponIdx = idx
				}
			}

			botWeaponIdx := 1

			if rand.Intn(1) == 0 || rules[game.LastBotWeaponIdx][game.LastUserWeaponIdx] == 0 {
				botWeaponIdx = rand.Intn(len(weapons))
			} else {
				if rules[game.LastBotWeaponIdx][game.LastUserWeaponIdx] > 0 {
					//Bot won
					botWeaponIdx = game.LastBotWeaponIdx

				} else if rules[game.LastBotWeaponIdx][game.LastUserWeaponIdx] < 0 {
					//User won
					for idx := range weapons {
						if idx != game.LastBotWeaponIdx && idx != game.LastUserWeaponIdx {
							botWeaponIdx = idx
						}
					}
				}
			}

			botWeapon := weapons[botWeaponIdx]

			game.LastBotWeaponIdx = botWeaponIdx
			game.LastUserWeaponIdx = userWeaponIdx

			err = r.Text(ctx, botWeapon)
			if err != nil {
				log.Infof(ctx, "Text1 - Err:%s", err)
			}

			//result := "We can't stay on an equality!"
			if rules[botWeaponIdx][userWeaponIdx] > 0 {
				//Bot win
				//result = "Nice try, little boy"
				game.BotWin++
			} else if rules[botWeaponIdx][userWeaponIdx] < 0 {
				//User win
				//result = "A last one, please!"
				game.UserWin++
			}

			/*err = r.Text(ctx, result)
			if err != nil {
				log.Infof(ctx, "Text2 - Err:%s", err)
			}*/

			err = r.Text(ctx, fmt.Sprintf("%s : %d - %d : Bot", p.FirstName, game.UserWin, game.BotWin))
			if err != nil {
				log.Infof(ctx, "Text2 - Err:%s", err)
			}

			if !game.IsWon && game.UserWin == 5 && game.BotWin < 5 {

				game.IsWon = true

				err = r.Text(ctx, "You won, congrats 🏆.")
				if err != nil {
					log.Infof(ctx, "Text3 - Err:%s", err)
				}

				err = r.Text(ctx, "You can continue to play for the fun or enter `reset` to restart the game.")
				if err != nil {
					log.Infof(ctx, "Text4 - Err:%s", err)
				}

			} else if !game.IsWon && game.UserWin < 5 && game.BotWin == 5 {

				game.IsWon = true

				err = r.Text(ctx, "I'm the winner, sorry.")
				if err != nil {
					log.Infof(ctx, "Text5 - Err:%s", err)
				}

				err = r.Text(ctx, "You can continue to play for the fun or enter `reset` to restart the game.")
				if err != nil {
					log.Infof(ctx, "Text6 - Err:%s", err)
				}

			} else if math.Mod(float64(game.UserWin), 5.0) == 0.0 || math.Mod(float64(game.BotWin), 5.0) == 0.0 {

				err = r.Text(ctx, "You can continue to play for the fun or enter `reset` to restart the game.")
				if err != nil {
					log.Infof(ctx, "Text6 - Err:%s", err)
				}
			}

			err = r.TextWithReplies(ctx, "Choose your weapon", replies)
			if err != nil {
				log.Infof(ctx, "TextWithReplies1 - Err:%s", err)
			}

		} else if strings.ToLower(m.Text) == "reset" {

			//Reset the game
			game = Game{UserId: m.Sender.ID}

			err = r.Text(ctx, "Let's go for a new game.")
			if err != nil {
				log.Infof(ctx, "Text7 - Err:%s", err)
			}

			err = r.TextWithReplies(ctx, "Choose your weapon", replies)
			if err != nil {
				log.Infof(ctx, "TextWithReplies2 - Err:%s", err)
			}

		} else if game.UserWin == 0 && game.BotWin == 0 {

			err = r.Text(ctx, "Welcome in \"Choose Your weapon\" game. The first to 5 win the game.")
			if err != nil {
				log.Infof(ctx, "Text8 - Err:%s", err)
			}

			err = r.TextWithReplies(ctx, "Choose your weapon", replies)
			if err != nil {
				log.Infof(ctx, "TextWithReplies3 - Err:%s", err)
			}

		} else {

			err = r.Text(ctx, fmt.Sprintf("Are you serious %s ?", p.FirstName))
			if err != nil {
				log.Infof(ctx, "TextWithReplies4 - Err:%s", err)
			}

			err = r.TextWithReplies(ctx, "Choose your weapon", replies)
			if err != nil {
				log.Infof(ctx, "TextWithReplies5 - Err:%s", err)
			}

		}

		_, err = datastore.Put(ctx, gameKey, &game)
		if err != nil {
			log.Infof(ctx, "Put - Err:%s", err)
		}
	})

	http.Handle("/", client.Handler())
}
