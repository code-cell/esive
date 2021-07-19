package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	esive_grpc "github.com/code-cell/esive/grpc"
	"github.com/code-cell/esive/tick"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var bots = flag.Int("bots", 10, "")

func main() {
	flag.Parse()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	for i := 0; i < *bots; i++ {
		os.Stdout.Write([]byte("."))
		runBot(i)
	}
	fmt.Println("All bots running.")

	<-sigs
}

func runBot(n int) {
	name := fmt.Sprintf("Bot %d", n)

	var t *tick.Tick
	conn, err := grpc.Dial("localhost:9000",
		grpc.WithInsecure(),
		grpc.WithChainUnaryInterceptor(
			func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
				if t != nil {
					ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("tick", strconv.FormatInt(t.Current(), 10)))
				}
				return invoker(ctx, method, req, reply, cc, opts...)
			}),
	)
	if err != nil {
		panic(err)
	}

	client := esive_grpc.NewEsiveClient(conn)

	var md metadata.MD
	joinRes, err := client.Join(context.Background(), &esive_grpc.JoinReq{
		Name: name,
	}, grpc.Header(&md))
	if err != nil {
		panic(err)
	}
	serverTick := getTickFromMD(md)
	t = tick.NewTick(serverTick+3, time.Duration(joinRes.TickMilliseconds)*time.Millisecond)
	go t.Start()

	visRes, err := client.TickUpdates(context.Background(), &esive_grpc.TickUpdatesReq{})
	if err != nil {
		panic(err)
	}
	chatRes, err := client.ChatUpdates(context.Background(), &esive_grpc.ChatUpdatesReq{})
	if err != nil {
		panic(err)
	}
	go handleBot(client, name, visRes, chatRes)
}

func handleBot(client esive_grpc.EsiveClient, name string, visRes esive_grpc.Esive_TickUpdatesClient, chatRes esive_grpc.Esive_ChatUpdatesClient) {
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

	for {
		time.Sleep(time.Duration(rand.Intn(7)+3) * time.Second)
		x := rand.Int63n(3) - 1
		y := rand.Int63n(3) - 1
		_, err := client.SetVelocity(context.Background(), &esive_grpc.Velocity{X: x, Y: y})
		if err != nil {
			panic(err)
		}
	}
}

func getTickFromMD(md metadata.MD) int64 {
	str, found := md["tick"]
	if !found || len(str) == 0 {
		panic("Invalid tick from the server metadata")
	}
	serverTick, err := strconv.ParseInt(str[0], 10, 64)
	if err != nil {
		panic(err)
	}
	return serverTick
}
