package main

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/liftbridge-io/go-liftbridge"
	"github.com/nats-io/go-nats"
	"golang.org/x/net/context"
)

const count = 5

func main() {
	if err := createStream(); err != nil {
		panic(err)
	}
	conn, err := nats.DefaultOptions.Connect()
	if err != nil {
		panic(err)
	}
	defer conn.Flush()
	defer conn.Close()

	ackInbox := "acks"
	var wg sync.WaitGroup

	sub, err := conn.Subscribe(ackInbox, func(m *nats.Msg) {
		ack, err := liftbridge.UnmarshalAck(m.Data)
		if err != nil {
			panic(err)
		}
		fmt.Println("ack:", ack.StreamSubject, ack.StreamName, ack.Offset, ack.MsgSubject)
		wg.Done()
	})
	if err != nil {
		panic(err)
	}
	defer sub.Unsubscribe()

	wg.Add(count)
	fmt.Println("publishing")
	for i := 0; i < count; i++ {
		m := liftbridge.NewMessage([]byte(strconv.Itoa(i)),
			liftbridge.MessageOptions{Key: []byte("test"), AckInbox: ackInbox})
		if err := conn.Publish("bar", m); err != nil {
			panic(err)
		}
	}
	fmt.Println("done publishing")

	wg.Wait()
}

func createStream() error {
	addr := "localhost:9292"
	client, err := liftbridge.Connect([]string{addr})
	if err != nil {
		return err
	}
	defer client.Close()
	stream := liftbridge.StreamInfo{
		Subject:           "bar",
		Name:              "bar-stream",
		ReplicationFactor: 1,
	}
	if err := client.CreateStream(context.Background(), stream); err != nil {
		if err != liftbridge.ErrStreamExists {
			return err
		}
	}
	fmt.Println("created stream foo-stream")
	return nil
}
