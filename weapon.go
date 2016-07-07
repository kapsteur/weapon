package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/kapsteur/messenger"
	"github.com/vrischmann/envconfig"
)

type Config struct {
	VerifyToken string `envconfig:"VERIFY_TOKEN"`
	Verify      bool   `envconfig:"VERIFY,optional"`
	PageToken   string `envconfig:"PAGE_TOKEN"`
}

var (
	conf    *Config
	weapons = []string{"✌", "✊", "✋"}
	replies = []messenger.QuickReplie{
		messenger.QuickReplie{ContentType: "text", Title: "✌", Payload: "1"},
		messenger.QuickReplie{ContentType: "text", Title: "✊", Payload: "2"},
		messenger.QuickReplie{ContentType: "text", Title: "✋", Payload: "3"}}
	rules = [][]int{[]int{0, -1, 1}, []int{1, 0, -1}, []int{-1, 1, 0}}
)

func main() {
	if err := envconfig.Init(&conf); err != nil {
		log.Fatal("err=%s\n", err)
	}

	// Create a new messenger client
	client := messenger.New(messenger.Options{
		Verify:      conf.Verify,
		VerifyToken: conf.VerifyToken,
		Token:       conf.PageToken,
	})

	// Setup a handler to be triggered when a message is received
	client.HandleMessage(func(m messenger.Message, r *messenger.Response) {
		rand.Seed(time.Now().UnixNano())

		log.Printf("%v (Sent, %v)", m.Text, m.Time.Format(time.UnixDate))

		p, err := client.ProfileByID(m.Sender.ID)
		if err != nil {
			log.Printf("ProfileByID - Err:%s", err)
		}

		if m.Text == "✌" || m.Text == "✊" || m.Text == "✋" {

			userWeaponIdx := 0
			for idx, w := range weapons {
				if w == m.Text {
					userWeaponIdx = idx
				}
			}

			botWeapon := "✊"

			//if rand.Intn(1) == 0 {
			botWeaponIdx := rand.Intn(len(weapons))
			botWeapon = weapons[botWeaponIdx]
			//} else {

			//}

			err = r.Text(botWeapon)
			if err != nil {
				log.Printf("Text1 - Err:%s", err)
			}

			result := "We can't stay on an equality!"
			if rules[botWeaponIdx][userWeaponIdx] > 0 {
				//Bot win
				result = "Nice try, little boy"
			} else if rules[botWeaponIdx][userWeaponIdx] < 0 {
				//User win
				result = "A last one, please!"
			}

			err = r.Text(result)
			if err != nil {
				log.Printf("Text2 - Err:%s", err)
			}

			err = r.TextWithReplies("Choose your weapon", &replies)
			if err != nil {
				log.Printf("TextWithReplies1 - Err:%s", err)
			}
		} else {
			err = r.TextWithReplies(fmt.Sprintf("Are you serious %s ?", p.FirstName), &replies)
			if err != nil {
				log.Println("TextWithReplies2 - Err:%s", err)
			}
		}
	})

	/*// Setup a handler to be triggered when a message is read
	client.HandleDelivery(func(d messenger.Delivery, r *messenger.Response) {
		log.Println(d.Watermark().Format(time.UnixDate))
	})*/

	log.Println("Serving messenger bot on localhost:7823")

	log.Fatal(http.ListenAndServe("localhost:7823", client.Handler()))
}
