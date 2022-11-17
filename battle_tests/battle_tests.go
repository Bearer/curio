package main

import (
	"github.com/rs/zerolog/log"
	"golang.org/x/net/context"

	"os"
	"os/signal"
	"syscall"

	"github.com/bearer/curio/battle_tests/config"
	"github.com/bearer/curio/battle_tests/db"
	"github.com/bearer/curio/battle_tests/rediscli"
	"github.com/bearer/curio/battle_tests/sheet"
	"github.com/bearer/curio/battle_tests/sync"
	"github.com/bearer/curio/cmd/curio/build"
)

type CurioVersion struct {
	Version string
}

func main() {
	config.Load()
	rediscli.Setup()
	err := rediscli.Init()

	if err != nil {
		log.Debug().Msgf("failed to init redis")
	}

	log.Printf("version %s", build.Version)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)
	signal.Notify(shutdown, syscall.SIGTERM)

	db := db.UnmarshalRaw()
	sheetClient := sheet.New()

	programCtx := context.TODO()
	docID, err := sync.GetDocumentID(sheetClient)
	if err != nil {
		log.Debug().Msgf("failed to get document id")
		log.Err(err).Send()
		return
	}

	log.Debug().Msgf("DocID is %s", docID)

	workerCtx := sync.DoWork(programCtx, db, docID, sheetClient)

	select {
	case <-shutdown:
		workerCtx.Done()
	case <-workerCtx.Done():
		return
	}
}
