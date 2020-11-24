package geolocate

import (
	"context"
	"net/http"

	"github.com/ooni/probe-engine/model"
	"github.com/pion/stun"
)

// STUNClient is the STUN client expected by this package
type STUNClient interface {
	Close() error
	Start(m *stun.Message, h stun.Handler) error
}

// STUNConfig contains configuration for STUNIPLookup
type STUNConfig struct {
	Dial     func(network string, address string) (STUNClient, error)
	Endpoint string
	Logger   model.Logger
}

func stundialer(network string, address string) (STUNClient, error) {
	return stun.Dial(network, address)
}

// STUNIPLookup performs the IP lookup using STUN.
func STUNIPLookup(ctx context.Context, config STUNConfig) (string, error) {
	config.Logger.Debugf("STUNIPLookup: start using %s", config.Endpoint)
	ip, err := func() (string, error) {
		dial := config.Dial
		if dial == nil {
			dial = stundialer
		}
		clnt, err := dial("udp", config.Endpoint)
		if err != nil {
			return model.DefaultProbeIP, err
		}
		defer clnt.Close()
		message := stun.MustBuild(stun.TransactionID, stun.BindingRequest)
		errch, ipch := make(chan error, 1), make(chan string, 1)
		err = clnt.Start(message, func(ev stun.Event) {
			if ev.Error != nil {
				errch <- ev.Error
				return
			}
			var xorAddr stun.XORMappedAddress
			if err := xorAddr.GetFrom(ev.Message); err != nil {
				errch <- err
				return
			}
			ipch <- xorAddr.IP.String()
		})
		if err != nil {
			return model.DefaultProbeIP, err
		}
		select {
		case err := <-errch:
			return model.DefaultProbeIP, err
		case ip := <-ipch:
			return ip, nil
		case <-ctx.Done():
			return model.DefaultProbeIP, ctx.Err()
		}
	}()
	if err != nil {
		config.Logger.Debugf("STUNIPLookup: failure using %s: %+v", config.Endpoint, err)
		return model.DefaultProbeIP, err
	}
	return ip, nil
}

// STUNEkigaIPLookup performs the IP lookup using ekiga.net.
func STUNEkigaIPLookup(
	ctx context.Context,
	httpClient *http.Client,
	logger model.Logger,
	userAgent string,
) (string, error) {
	return STUNIPLookup(ctx, STUNConfig{
		Endpoint: "stun.ekiga.net:3478",
		Logger:   logger,
	})
}

// STUNGoogleIPLookup performs the IP lookup using google.com.
func STUNGoogleIPLookup(
	ctx context.Context,
	httpClient *http.Client,
	logger model.Logger,
	userAgent string,
) (string, error) {
	return STUNIPLookup(ctx, STUNConfig{
		Endpoint: "stun.l.google.com:19302",
		Logger:   logger,
	})
}
