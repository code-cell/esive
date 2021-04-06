package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"time"

	"github.com/code-cell/esive/models"
	"google.golang.org/grpc"
)

var bots = flag.Int("bots", 1000, "")

func main() {
	flag.Parse()

	for i := 0; i < *bots; i++ {
		runBot(i)
	}

	time.Sleep(200000 * time.Second)
}

func runBot(n int) {
	name := fmt.Sprintf("Bot %d", n)
	conn, err := grpc.Dial("localhost:9000", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	client := models.NewIcecreamClient(conn)

	_, err = client.Join(context.Background(), &models.JoinReq{
		Name: name,
	})
	if err != nil {
		panic(err)
	}

	visRes, err := client.VisibilityUpdates(context.Background(), &models.VisibilityUpdatesReq{})
	if err != nil {
		panic(err)
	}
	chatRes, err := client.ChatUpdates(context.Background(), &models.ChatUpdatesReq{})
	if err != nil {
		panic(err)
	}
	go handleBot(client, name, visRes, chatRes)
}

func handleBot(client models.IcecreamClient, name string, visRes models.Icecream_VisibilityUpdatesClient, chatRes models.Icecream_ChatUpdatesClient) {
	go func() {

		for {
			_, err := visRes.Recv()
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		}
	}()
	go func() {
		for {
			msg, err := chatRes.Recv()
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			if msg.Message.Text == "Hello" {
				_, err := client.Say(context.Background(), &models.SayReq{
					Text: fmt.Sprintf("Hello %v, my name is %v", msg.Message.From, name),
				})
				if err != nil {
					panic(err)
				}
			}
		}
	}()

	time.Sleep(10000 * time.Second)
	for {
		time.Sleep(time.Duration(rand.Intn(1000)+1000) * time.Millisecond)
		switch rand.Intn(4) {
		case 0:
			_, err := client.MoveUp(context.Background(), &models.MoveUpReq{})
			if err != nil {
				panic(err)
			}
		case 1:
			_, err := client.MoveDown(context.Background(), &models.MoveDownReq{})
			if err != nil {
				panic(err)
			}
		case 2:
			_, err := client.MoveLeft(context.Background(), &models.MoveLeftReq{})
			if err != nil {
				panic(err)
			}
		case 3:
			_, err := client.MoveRight(context.Background(), &models.MoveRightReq{})
			if err != nil {
				panic(err)
			}
		}
	}
}
