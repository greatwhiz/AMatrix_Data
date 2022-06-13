package main

import (
	"A-Matrix/src/binance"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

func init() {
	setLog()
}

func main() {
	binance.UpdateSymbols()
	binance.UpdateArbitrageRelation()
	binance.SubscribeMarket()
}

func setLog() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}
	dir := filepath.Dir(filename)
	f, err := os.OpenFile(fmt.Sprintf("%s/../../tmp/log.txt", dir), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
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
