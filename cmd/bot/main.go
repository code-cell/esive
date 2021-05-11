package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"time"

	esive_grpc "github.com/code-cell/esive/grpc"
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

	client := esive_grpc.NewEsiveClient(conn)

	_, err = client.Join(context.Background(), &esive_grpc.JoinReq{
		Name: name,
	})
	if err != nil {
		panic(err)
	}

	visRes, err := client.VisibilityUpdates(context.Background(), &esive_grpc.VisibilityUpdatesReq{})
	if err != nil {
		panic(err)
	}
	chatRes, err := client.ChatUpdates(context.Background(), &esive_grpc.ChatUpdatesReq{})
	if err != nil {
		panic(err)
	}
	go handleBot(client, name, visRes, chatRes)
}

func handleBot(client esive_grpc.EsiveClient, name string, visRes esive_grpc.Esive_VisibilityUpdatesClient, chatRes esive_grpc.Esive_ChatUpdatesClient) {
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
				_, err := client.Say(context.Background(), &esive_grpc.SayReq{
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
		time.Sleep(time.Duration(rand.Intn(10000)+1000) * time.Millisecond)
		x := rand.Int63n(3) - 1
		y := rand.Int63n(3) - 1
		_, err := client.SetVelocity(context.Background(), &esive_grpc.Velocity{X: x, Y: y})
		if err != nil {
			panic(err)
		}
	}
}
