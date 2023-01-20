package loader

import (
	"context"
	"log"
	"reflect"
	"runtime"

	"github.com/IlliniBlockchain/etl-bitcoin/client"
	"github.com/IlliniBlockchain/etl-bitcoin/database"
)

// LoaderMsg stores state for data being passed through loaders.
type LoaderMsg[T any] struct {
	dbTx *dBTxWithStats
	data T
}

// ILoader is a simple interface for loaders.
type ILoader interface {
	Run(ctx context.Context) error
}

// Loader is the go to loader for an inidividual stage in the pipeline
// extracting data from an RPC client to a database. It stores state and
// uses the function f to transform data coming from a src channel to send
// to a dst channel.
type Loader[S, D any] struct {
	client client.Client
	src    <-chan *LoaderMsg[S]
	dst    chan *LoaderMsg[D]
	f      LoaderFunc[S, D]
}

type LoaderFunc[S, D any] func(client.Client, S) (D, error)

func (loader *Loader[S, D]) Dst() <-chan *LoaderMsg[D] {
	return loader.dst
}

func NewLoader[S, D any](client client.Client, src <-chan *LoaderMsg[S], f LoaderFunc[S, D]) *Loader[S, D] {
	dst := make(chan *LoaderMsg[D], 1)
	loader := &Loader[S, D]{
		client,
		src,
		dst,
		f,
	}
	return loader
}

// Run listens for messages sent to the loader's src channel,
// transforms the data and sends it to the next loader.
// Importantly, when it's src channel is closed, it closes its
// dst channel, which causes a domino effect of closing loader channels.
func (loader *Loader[S, D]) Run(ctx context.Context) error {
	// Close dst to signal to downstream loaders that there is
	// no more data whenever execution stops.
	defer close(loader.dst)

	// Listen for messages from upstream loaders.
	for msg := range loader.src {
		log.Printf("%s (%s) started\n", runtime.FuncForPC(reflect.ValueOf(loader.f).Pointer()).Name(), msg.dbTx.blockRange)
		output, err := loader.f(loader.client, msg.data)
		log.Printf("%s (%s) finished\n", runtime.FuncForPC(reflect.ValueOf(loader.f).Pointer()).Name(), msg.dbTx.blockRange)
		if err != nil {
			return err
		}
		loader.dst <- &LoaderMsg[D]{msg.dbTx, output}
		log.Printf("%s (%s) sent\n", runtime.FuncForPC(reflect.ValueOf(loader.f).Pointer()).Name(), msg.dbTx.blockRange)
	}
	return nil
}

// LoaderSink represents the last stage in a loader pipeline.
type LoaderSink[S any] struct {
	src <-chan *LoaderMsg[S]
	f   LoaderSinkFunc[S]
}

type LoaderSinkFunc[S any] func(database.DBTx, S) error

func NewLoaderSink[S any](src <-chan *LoaderMsg[S], f LoaderSinkFunc[S]) *LoaderSink[S] {
	loader := &LoaderSink[S]{
		src,
		f,
	}
	return loader
}

func (loader *LoaderSink[S]) Run(ctx context.Context) error {
	for msg := range loader.src {
		log.Printf("%s (%s) started\n", runtime.FuncForPC(reflect.ValueOf(loader.f).Pointer()).Name(), msg.dbTx.blockRange)
		if err := loader.f(msg.dbTx, msg.data); err != nil {
			return err
		}
		if err := msg.dbTx.Commit(); err != nil {
			return err
		}
	}
	return nil
}
