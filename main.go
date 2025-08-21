package main

import (
	"context"
	"log"

	"go-esb-store/pkg/trigger"
)

func main() {
	log.Println("Starting function locally...")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	e := &trigger.LocalEvent{Body: string(trigger.LocalSource)}

	res, err := Handler(ctx, e)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Local function finished successfully")
	log.Println(res)
}
