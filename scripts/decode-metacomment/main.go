package main

import (
	"fmt"
	"log"
	"os"

	"github.com/reviewdog/reviewdog/service/commentutil"
)

func main() {
	if len(os.Args) == 1 {
		log.Fatal("require one argument")
	}
	meta, err := commentutil.DecodeMetaComment(os.Args[1])
	if err != nil {
		log.Fatalf("failed to decode meta comment: %v", err)
	}
	fmt.Printf("%v\n", meta)
}
