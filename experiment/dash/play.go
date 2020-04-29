package dash

import (
	"context"
	"fmt"
	"time"

	"github.com/ooni/probe-engine/internal/humanizex"
)

type playConfig struct {
	authorization string
	fqdn          string
	numIterations int64
	realAddress   string
}

func (r runner) play(ctx context.Context, config playConfig) error {
	errch := make(chan error)
	playoutbuf := make(chan clientResults, 1) // strive to keep one in buffer
	go fetch(ctx, fetchConfig{
		authorization: config.authorization,
		deps:          r,
		errch:         errch,
		fqdn:          config.fqdn,
		numIterations: config.numIterations,
		outch:         playoutbuf,
		realAddress:   config.realAddress,
		saver:         r.saver,
	})
	// get the first frame
	var (
		err   error
		frame clientResults
	)
	select {
	case err = <-errch:
		return err
	case frame = <-playoutbuf:
	}
	for {
		// record the current frame
		r.tk.ReceiverData = append(r.tk.ReceiverData, frame)
		// play the current frame
		percentage := float64(frame.Iteration) / float64(config.numIterations)
		rate := 8 * float64(frame.Received) / float64(frame.ElapsedTarget)
		msg := fmt.Sprintf("streaming: play at %s", humanizex.SI(rate, "bit/s"))
		r.callbacks.OnProgress(percentage, msg)
		<-time.After(time.Duration(frame.ElapsedTarget) * time.Second)
		// get the next frame nonblocking
		select {
		case err = <-errch:
			return err
		case frame = <-playoutbuf:
		default:
			// get the next frame blocking
			frame, err = r.playStall(percentage, playoutbuf, errch)
			if err != nil {
				return err
			}
		}
	}
}

func (r runner) playStall(percentage float64,
	playoutbuf chan clientResults, errch chan error) (clientResults, error) {
	progress := func(stall float64) {
		msg := fmt.Sprintf("streaming: stalled for %.1f s", stall)
		r.callbacks.OnProgress(percentage, msg)
	}
	begin := time.Now()
	for {
		select {
		case err := <-errch:
			return clientResults{}, err
		case frame := <-playoutbuf:
			stall := time.Now().Sub(begin).Seconds()
			if stall > r.tk.Simple.MinPlayoutDelay {
				r.tk.Simple.MinPlayoutDelay = stall
			}
			progress(stall)
			return frame, nil
		case now := <-time.After(1 * time.Second):
			progress(now.Sub(begin).Seconds())
		}
	}
}
