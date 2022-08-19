package service

import (
	"fmt"
	"io"
	"sync"
	"time"

	pb "event-emitter/events"
)


type EventsServer struct {
	pb.UnimplementedEventsServer
	eventClients sync.Map
}

type eventClient struct {
	stream   pb.Events_DeliverMPTEventServer // stream is the server side of the RPC stream
	finished chan<- bool                     // used to signal closure of a client subscribing goroutine
}

func (s *EventsServer) DeliverMPTEvent(stream pb.Events_DeliverMPTEventServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			fmt.Println("Error in DeliverMPT line 39: ", err)
			return err
		}

		fmt.Println("DeliverMPTEvent called by Client : (id)", in.ClientId)

		fin := make(chan bool)

		s.eventClients.Store(in.ClientId, eventClient{stream: stream, finished: fin})

		ctx := stream.Context()

		for {
			select {
			case <-fin:
				fmt.Printf("Closing stream for client ID: %s", in.ClientId)
				return nil
			case <- ctx.Done():
				fmt.Printf("Client ID %s has disconnected", in.ClientId)
				return nil
			}
		}
	}
}

type EventType string

func (s *EventsServer) Emit(eventType EventType, mptEvent *pb.MPTEventResponse) {
	time.Sleep(time.Second * 5)
	for {
	   var unsubscribe []string
 
	   // Iterate over all subscribers and send data to each client
	   s.eventClients.Range(func(k, v interface{}) bool {
			id, ok := k.(string)
			if !ok {
				fmt.Printf("Failed to cast subscriber key: %T", k)
				return false
			}

			ec, ok := v.(eventClient)
			if !ok {
				fmt.Printf("Failed to cast subscriber value: %T", v)
				return false
			}

			fmt.Println("Sending data over the gRPC stream to the client: ", id)

			if err := ec.stream.Send(mptEvent); err != nil {

				fmt.Printf("Failed to send data to client: %v", err)
				select {
					case ec.finished <- true:
						fmt.Printf("Unsubscribed client: %s", id)
					default:
					// Default case is to avoid blocking in case client has already unsubscribed
				}
				// In case of error the client would re-subscribe so close the subscriber stream
				unsubscribe = append(unsubscribe, id)
		  }

		  return true
	   })
 
	   // Remove erroneous client streams
	   for _, id := range unsubscribe {
		  s.eventClients.Delete(id)
	   }
	}
 }