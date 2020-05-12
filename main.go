package main

import (
	"fmt"
	"log"
	"os"

	"github.com/go-generalize/pubsub_gen/deploy"
	"github.com/go-generalize/pubsub_gen/migrate"

	"github.com/go-generalize/pubsub_gen/boilerplate"

	pubsubgogen "github.com/go-generalize/pubsub_gen/gogen"
)

func main() {
	if len(os.Args) < 2 {
		cmd := os.Args[0]
		fmt.Printf("%s: Code Generator for Cloud PubSub\n\n", cmd)

		fmt.Printf("Subcommands:\n")
		fmt.Printf("\t%s boilerplate: Generate boilerplate for Cloud PubSub\n", cmd)
		fmt.Printf("\t%s gogen:       Generate struct/interface of publisher/subscriber(used by go generate)\n", cmd)
		fmt.Printf("\t%s migrate:     Create insufficient topics(mainly used in CI/CD)\n", cmd)
		fmt.Printf("\t%s deploy:      Deploy Cloud Functions\n", cmd)

		return
	}

	sub := os.Args[1]
	os.Args = append([]string{os.Args[0]}, os.Args[2:]...)

	switch sub {
	case "boilerplate":
		boilerplate.Main()
	case "gogen":
		pubsubgogen.Main()
	case "migrate":
		migrate.Main()
	case "deploy":
		deploy.Main()
	default:
		log.Println("unknown sub command: ", sub)
	}
}
