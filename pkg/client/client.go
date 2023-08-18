package client

import (
	context "context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"time"

	pb "github.com/tauraamui/bluepanda/pkg/api"
	"github.com/tauraamui/bluepanda/pkg/kvs"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var addr = flag.String("addr", "localhost:50051", "the address to connect to")

func clientstub() {
	flag.Parse()
	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewBluePandaClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	rootUUID := (kvs.RootOwner{}).String()
	r, err := c.Fetch(ctx, &pb.FetchRequest{Type: "fruit", Uuid: rootUUID, Columns: []string{"name"}})
	if err != nil {
		log.Fatalf("could not fetch: %v", err)
	}

	for {
		rslt, err := r.Recv()
		if err == io.EOF {
			break
		}

		data := map[string]any{}
		if err := json.Unmarshal(rslt.Json, &data); err != nil {
			log.Fatalf("could not unmarshal response data: %v", err)
		}

		fmt.Printf("%+v\n", data)
	}
}
