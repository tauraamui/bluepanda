package client

import (
	context "context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	pb "github.com/tauraamui/bluepanda/pkg/api"
	"github.com/tauraamui/bluepanda/pkg/kvs"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type fruit struct {
	Name string
}

func Run() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(":3000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewBluePandaClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	newFruit := fruit{Name: "dragonfruit"}
	body, err := json.Marshal(newFruit)
	if err != nil {
		log.Fatalf("unable to encode fruit to send: %v", err)
	}
	rootUUID := (kvs.RootOwner{}).String()
	insertResult, err := c.Insert(ctx, &pb.InsertRequest{Type: "fruit", Uuid: rootUUID, Json: body})
	if err != nil {
		log.Fatalf("failed to insert entry: %v", err)
	}

	fmt.Printf("insert result: %s\n", insertResult.Status)

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
