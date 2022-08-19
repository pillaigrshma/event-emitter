package main

import (
	"context"
	pb "event-emitter/events"
	"fmt"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
)

func main() {
	// Create multiple clients and start receiving data
	var wg sync.WaitGroup

	for i := 1; i <= 3; i++ {
		wg.Add(1)
		client, err := mkDeliverClient(fmt.Sprint(i))
		if err != nil {
			log.Fatal(err)
		}
		// Dispatch client goroutine
		go client.start()
		time.Sleep(time.Second * 2)
	}

	// The wait group purpose is to avoid exiting, the clients do not exit
	wg.Wait()
}

// EventsClient holds the long lived gRPC client fields
type EventsClient struct {
	client pb.EventsClient  // client is the long lived gRPC client
	conn   *grpc.ClientConn // conn is the client gRPC connection
	id     string           // id is the client ID used for subscribing
}

// mkDeliverClient creates a new client instance
func mkDeliverClient(id string) (*EventsClient, error) {
	conn, err := mkConnection()
	if err != nil {
		return nil, err
	}
	return &EventsClient{
		client: pb.NewEventsClient(conn),
		conn:   conn,
		id:     id,
	}, nil
}

// close the gRPC client connection
func (c *EventsClient) close() {
	if err := c.conn.Close(); err != nil {
		log.Fatal(err)
	}
}

// subscribe subscribes to messages from the gRPC server
func (c *EventsClient) subscribe() (pb.Events_DeliverMPTEventClient, error) {
	log.Printf("Subscribing client ID: %s", c.id)
	return c.client.DeliverMPTEvent(context.Background())
}

func (c *EventsClient) start() {
	var err error
	// stream is the client side of the RPC stream
	var stream pb.Events_DeliverMPTEventClient
	for {
		if stream == nil {
			if stream, err = c.subscribe(); err != nil {
				log.Printf("Failed to subscribe: %v", err)
				c.sleep()
				// Retry on failure
				continue
			}
			envelope := &pb.MPTEventRequest{
				ClientId:  c.id,
				EventType: "info",
			}
			if err := stream.Send(envelope); err != nil {
				log.Fatalf("client.DeliverMPTEvent: stream.Send(%v) failed: %v", envelope, err)
			}

		}

		response, err := stream.Recv()
		if err != nil {
			log.Printf("Failed to receive message: %v", err)
			// Clearing the stream will force the client to resubscribe on next iteration
			stream = nil
			c.sleep()
			// Retry on failure
			continue
		}

		log.Printf("Client ID %s got response: %q", c.id, response.Message)
	}
}

// sleep is used to give the server time to unsubscribe the client and reset the stream
func (c *EventsClient) sleep() {
	time.Sleep(time.Second * 5)
}

func mkConnection() (*grpc.ClientConn, error) {
	return grpc.Dial("127.0.0.1:50051", []grpc.DialOption{grpc.WithInsecure(), grpc.WithBlock()}...)
}
