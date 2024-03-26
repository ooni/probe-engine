package webconnectivitylte

import (
	"github.com/ooni/probe-engine/pkg/model"
	"github.com/ooni/probe-engine/pkg/webconnectivityalgo"
)

// DNSWhoamiSingleton is the DNSWhoamiService singleton.
var DNSWhoamiSingleton = webconnectivityalgo.NewDNSWhoamiService(model.DiscardLogger)
