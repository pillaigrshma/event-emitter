// package main

// import (
// 	"context"
// 	pb "event-emitter/events"
// 	"fmt"
// 	"log"
// 	"sync"
// 	"time"

// 	"google.golang.org/grpc"
// )

// func (c *EventsClient) StartTest() {
// 	var err error
// 	// stream is the client side of the RPC stream
// 	var stream pb.Events_DeliverMPTEventClient
// 	for {
// 		if stream == nil {
// 			if stream, err = c.subscribe(); err != nil {
// 				log.Printf("Failed to subscribe: %v", err)
// 				c.sleep()
// 				// Retry on failure
// 				continue
// 			}
// 			envelope := &pb.MPTEventRequest{
// 				ClientId:  c.id,
// 				EventType: "info",
// 			}
// 			if err := stream.Send(envelope); err != nil {
// 				log.Fatalf("client.DeliverMPTEvent: stream.Send(%v) failed: %v", envelope, err)
// 			}

// 		}

// 		response, err := stream.Recv()
// 		if err != nil {
// 			log.Printf("Failed to receive message: %v", err)
// 			// Clearing the stream will force the client to resubscribe on next iteration
// 			stream = nil
// 			c.sleep()
// 			// Retry on failure
// 			continue
// 		}

// 		log.Printf("Client ID %s got response: %q", c.id, response.Message)
// 	}
// }


// // func main() {

// // }