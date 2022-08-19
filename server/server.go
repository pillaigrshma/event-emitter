package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
	"io"
	"time"

	"google.golang.org/grpc"

	pb "event-emitter/events"
)

var (
	tls        = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	port       = flag.Int("port", 50051, "The server port")
)

type EventsServer struct {
	pb.UnimplementedEventsServer
	subscribers sync.Map
}

type subInfo struct {
	stream   pb.Events_DeliverMPTEventServer // stream is the server side of the RPC stream
	finished chan<- bool                 // used to signal closure of a client subscribing goroutine
}

func (s *EventsServer) DeliverMPTEvent(stream pb.Events_DeliverMPTEventServer) error {
	//for {
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

	s.subscribers.Store(in.ClientId, subInfo{stream: stream, finished: fin})

	ctx := stream.Context()

	for {
		select {
		case <-fin:
			log.Printf("Closing stream for client ID: %s", in.ClientId)
			return nil
		case <- ctx.Done():
			log.Printf("Client ID %s has disconnected", in.ClientId)
			return nil
		}
	}
	//}
	
}

func (s *EventsServer) mockDataGenerator() {
	log.Println("Server is running...")
	for {
	   time.Sleep(time.Second)
 
	   // A list of clients to unsubscribe in case of error
	   var unsubscribe []string
 
	   // Iterate over all subscribers and send data to each client
	   s.subscribers.Range(func(k, v interface{}) bool {
			id, ok := k.(string)
			if !ok {
				log.Printf("Failed to cast subscriber key: %T", k)
				return false
			}

			subInfo, ok := v.(subInfo)
			if !ok {
				log.Printf("Failed to cast subscriber value: %T", v)
				return false
			}

			fmt.Println("Sending data over the gRPC stream to the client: ", id)
			if err := subInfo.stream.Send(&pb.MPTEventResponse{
				Message: fmt.Sprintf("data message for client: %s", id),
			}); err != nil {

				log.Printf("Failed to send data to client: %v", err)
				select {
					case subInfo.finished <- true:
						log.Printf("Unsubscribed client: %s", id)
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
		  s.subscribers.Delete(id)
	   }
	}
 }
 
 

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	var opts []grpc.ServerOption

	if *tls {
		log.Fatalf("Failed to generate credentials %v", err)
	}

	newDeliverServer := &EventsServer{}

	go newDeliverServer.mockDataGenerator()

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterEventsServer(grpcServer, newDeliverServer)
	
	fmt.Println("Server listening ...")

	grpcServer.Serve(lis)

}
