package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"time"

	stan "github.com/nats-io/go-nats-streaming"

	"github.com/searchlight/auditsink-to-rbac/rbac"
	"github.com/searchlight/auditsink-to-rbac/system"
)

func ProcessMessage(msg *stan.Msg) {
	log.Println(string(msg.Data))
	println("")
	if err := rbac.SaveEventList(msg.Data); err != nil {
		log.Println(err)
	}

	if err := rbac.SaveAuditLogToDatabaseFromBytes(msg.Data); err != nil {
		log.Println(err)
	}

	if err := rbac.CreateRoleFromBytes(msg.Data); err != nil {
		log.Println(err)
	}
	msg.Ack()
}

func logCloser(closer io.Closer) {
	if err := closer.Close(); err != nil {
		log.Println("Close error:", err)
	}
}
func main() {
	conn, err := stan.Connect(
		system.ClusterID,
		system.SubClientID,
		stan.NatsURL(stan.DefaultNatsURL),
		stan.ConnectWait(10*time.Second),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			log.Fatalln("Connection lost, reason:", reason)
		}),
	)
	if err != nil {
		log.Fatalf("Can't connect: %v.\nMake sure a NATS Streaming Server is running at: %s", err, stan.DefaultNatsURL)
	}
	defer logCloser(conn)

	rbac.StartXormEngine()
	defer rbac.EngineCloser()

	log.Printf("Connected to %s clusterID: [%s] clientID: [%s]\n", stan.DefaultNatsURL, system.ClusterID, system.SubClientID)

	sub, err := conn.QueueSubscribe(
		system.NatsSubject, system.SubscriberQueue,
		ProcessMessage, stan.SetManualAckMode(), stan.DurableName("i-remember"),
		stan.DeliverAllAvailable(), stan.AckWait(time.Second),
	)
	defer logCloser(sub)

	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt)
	<-channel
	println("")
	log.Println("Queue subscriber has been closed...")
}
