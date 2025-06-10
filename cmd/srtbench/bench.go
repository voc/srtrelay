package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	gosrt "github.com/datarhei/gosrt"
)

func main() {
	addr := flag.String("addr", "localhost:1337", "address to connect to")
	streams := flag.Int("streams", 1, "number of streams")
	ratio := flag.Int("ratio", 5, "number of subscribers per publisher")
	flag.Parse()
	var pubs []*pub
	var subs []*sub

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx, _ = signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	t := &test{
		addr:    *addr,
		errChan: make(chan error, 1),
		log:     gosrt.NewLogger([]string{"connection:error"}),
	}
	for i := range *streams {
		pub, err := t.runPublisher(ctx, fmt.Sprintf("publish/%d", i))
		if err != nil {
			log.Fatal("publisher:", err)
		}
		pubs = append(pubs, pub)
		for range *ratio {
			sub, err := t.runSubscriber(ctx, fmt.Sprintf("play/%d", i))
			if err != nil {
				log.Fatal("subscriber:", err)
			}
			subs = append(subs, sub)
		}
	}
	lastWrite := int64(0)
	lastRead := int64(0)
	for {
		select {
		case <-ctx.Done():
		case err := <-t.errChan:
			cancel()
			log.Println("error:", err)
		case log := <-t.log.Listen():
			fmt.Printf("[%s] %s %s:%d\n", log.Topic, log.Message, log.File, log.Line)
			continue
		case <-time.After(time.Second):
			writeTotal := int64(0)
			// var stats gosrt.Statistics
			for _, pub := range pubs {
				writeTotal += pub.written.Load()
				// pub.conn.Stats(&stats)
			}
			readTotal := int64(0)
			for _, sub := range subs {
				readTotal += sub.read.Load()
			}
			log.Printf("write: %.3fMbit/s, read: %.3fMbit/s", float64(writeTotal-lastWrite)/1e6*8, float64(readTotal-lastRead)/1e6*8)
			lastWrite = writeTotal
			lastRead = readTotal
			continue
		}
		break
	}
	t.done.Wait()
	log.Println("done")
}

type test struct {
	addr    string
	done    sync.WaitGroup
	errChan chan error
	log     gosrt.Logger
}

func (t *test) error(err error) {
	select {
	case t.errChan <- err:
	default:
	}
}

type pub struct {
	conn    gosrt.Conn
	written atomic.Int64
}

type sub struct {
	conn gosrt.Conn
	read atomic.Int64
}

// run single publisher sending to relay
func (t *test) runPublisher(ctx context.Context, streamid string) (*pub, error) {
	conf := gosrt.DefaultConfig()
	conf.StreamId = streamid
	conf.Latency = 50 * time.Millisecond
	conf.InputBW = 384000 // 3Mbit/s
	conf.MaxBW = 0
	conf.OverheadBW = 25
	conn, err := gosrt.Dial("srt", t.addr, conf)
	if err != nil {
		return nil, err
		// handle error
	}
	buffer := make([]byte, 1316)
	pub := &pub{
		conn: conn,
	}
	t.done.Add(2)
	context.AfterFunc(ctx, func() {
		defer t.done.Done()
		conn.Close()
	})

	go func() {
		defer t.done.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(3436 * time.Microsecond):
				// 1 packet every 3.436 milliseconds -> 3mbit/s
			}
			if err := conn.SetWriteDeadline(time.Now().Add(300 * time.Millisecond)); err != nil {
				t.error(fmt.Errorf("set write deadline failed: %w", err))
				break
			}
			n, err := conn.Write(buffer)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				// handle error
				t.error(fmt.Errorf("write failed: %w", err))
				break
			}
			// handle received data
			pub.written.Add(int64(n))
		}
	}()
	return pub, nil
}

func (t *test) runSubscriber(ctx context.Context, streamid string) (*sub, error) {
	conf := gosrt.DefaultConfig()
	conf.Latency = 50 * time.Millisecond
	conf.StreamId = streamid
	conf.PeerIdleTimeout = time.Second * 10
	conf.ConnectionTimeout = time.Second * 10
	conf.Logger = t.log
	conn, err := gosrt.Dial("srt", t.addr, conf)
	if err != nil {
		return nil, err
		// handle error
	}
	buffer := make([]byte, 2048)
	sub := &sub{
		conn: conn,
	}
	t.done.Add(2)
	context.AfterFunc(ctx, func() {
		defer t.done.Done()
		conn.Close()
	})

	go func() {
		defer t.done.Done()
		for {
			if ctx.Err() != nil {
				return
			}
			if err := conn.SetReadDeadline(time.Now().Add(300 * time.Millisecond)); err != nil {
				t.error(fmt.Errorf("set read deadline failed: %w", err))
				break
			}
			n, err := conn.Read(buffer)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				// handle error
				t.error(fmt.Errorf("read failed: %w", err))
				break
			}
			sub.read.Add(int64(n))
		}
	}()
	return sub, nil
}
