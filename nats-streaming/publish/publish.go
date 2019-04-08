package publish

import (
	"fmt"
	"io"
	"log"
	"time"

	stan "github.com/nats-io/go-nats-streaming"
	"github.com/searchlight/auditsink-to-rbac/system"
)

func logCloser(closer io.Closer) {
	if err := closer.Close(); err != nil {
		log.Println("Close error:", err)
	}
}

func PublishToNatsServer(msg []byte) error {
	conn, err := stan.Connect(
		system.ClusterID,
		system.PubClientID,
		stan.NatsURL(stan.DefaultNatsURL),
		stan.ConnectWait(2*time.Second),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			log.Fatalln("Connection lost, reason:", reason)
		}),
	)
	if err != nil {
		return fmt.Errorf("can't connect: %v.\nMake sure a NATS Streaming Server is running at: %s", err, stan.DefaultNatsURL)
	}
	defer logCloser(conn)

	log.Printf("Connected to %s clusterID: [%s] clientID: [%s]\n", stan.DefaultNatsURL, system.ClusterID, system.PubClientID)

	if err = conn.Publish(system.NatsSubject, msg); err != nil {
		return fmt.Errorf("error during AuditSink Event publishing: %s", err)
	}
	return nil
}
