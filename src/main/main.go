package main

import (
	"A-Matrix/src/binance_v1"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"log"
	"os"
	"path/filepath"
)

func init() {
	setLog()
}

func main() {
	binance_v1.UpdateSymbols()
	binance_v1.UpdateArbitrageRelation()
	binance_v1.SubscribeMarket()
}

func setLog() {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	f, err := os.OpenFile(fmt.Sprintf("%s/tmp/log.txt", exPath), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer func() {
		err := f.Close() // we must close anyway
		if err == nil {  // we must not overwrite the actual error if it is happened, and we did all the best to cleanup anyway
			err = errors.Wrap(err, "close")
		}
	}()

	wrt := io.MultiWriter(os.Stdout, f)
	log.SetOutput(wrt)
}
