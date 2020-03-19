package p2p

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	ma "github.com/multiformats/go-multiaddr"
)

func convertPeers(peers []string) ([]peer.AddrInfo, error) {
	pInfoS := make([]peer.AddrInfo, len(peers))
	for i, p := range peers {
		mAddr := ma.StringCast(p)
		pr, _ := peer.AddrInfoFromP2pAddr(mAddr)
		pInfoS[i] = *pr
	}
	return pInfoS, nil
}

// This code is borrowed from the go-ipfs bootstrap process
func bootstrapConnect(ctx context.Context, ph host.Host, peers []peer.AddrInfo) error {
	if len(peers) < 1 {
		return errors.New("not enough bootstrap peers")
	}

	errs := make(chan error, len(peers))
	var wg sync.WaitGroup
	for _, p := range peers {
		// performed asynchronously because when performed synchronously, if
		// one `Connect` call hangs, subsequent calls are more likely to
		// fail/abort due to an expiring context.
		// Also, performed asynchronously for dial speed.

		wg.Add(1)
		go func(p peer.AddrInfo) {
			defer wg.Done()
			//defer logger.Debug(ctx, "bootstrapDial", ph.ID(), p.ID)
			//logger.Debugf("%s bootstrapping to %s", ph.ID(), p.ID)

			ph.Peerstore().AddAddrs(p.ID, p.Addrs, peerstore.PermanentAddrTTL)
			if err := ph.Connect(ctx, p); err != nil {
				//logger.Errorf("Failed to bootstrap with %v: %s", p.ID, err)
				errs <- err
				return
			}
			//logger.Debug(ctx, "bootstrapDialSuccess", p.ID)
			//logger.Debugf("Bootstrapped with %v", p.ID)
		}(p)
	}
	wg.Wait()

	// our failure condition is when no connection attempt succeeded.
	// So drain the errs channel, counting the results.
	close(errs)
	count := 0
	var err error
	for err = range errs {
		if err != nil {
			count++
		}
	}
	if count == len(peers) {
		s := fmt.Sprintf("Failed to bootstrap. %s", err)
		return errors.New(s)
	}
	return nil
}
