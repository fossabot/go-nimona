package dht

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"go.uber.org/zap"

	"github.com/nimona/go-nimona/blocks"
	"github.com/nimona/go-nimona/log"
	"github.com/nimona/go-nimona/net"
	"github.com/nimona/go-nimona/peers"
)

var (
	ErrNotFound = errors.New("not found")
)

const (
	exchangeExtention    = "dht"
	closestPeersToReturn = 8
	maxQueryTime         = time.Second
)

// DHT is the struct that implements the dht protocol
type DHT struct {
	peerID         string
	store          *Store
	exchange       net.Exchange
	addressBook    *peers.AddressBook
	queries        sync.Map
	refreshBuckets bool
}

// NewDHT returns a new DHT from a exchange and peer manager
func NewDHT(exchange net.Exchange, pm *peers.AddressBook) (*DHT, error) {
	// create new kv store
	store, _ := newStore()

	// Create DHT node
	nd := &DHT{
		store:       store,
		exchange:    exchange,
		addressBook: pm,
		queries:     sync.Map{},
	}

	exchange.Handle("dht", nd.handleBlock)
	exchange.Handle(peers.PeerInfoType, nd.handleBlock)

	go nd.refresh()

	return nd, nil
}

func (nd *DHT) refresh() {
	ctx := context.Background()
	// TODO this will be replaced when we introduce bucketing
	// TODO our init process is a bit messed up and addressBook doesn't know
	// about the peer's protocols instantly
	for len(nd.addressBook.GetLocalPeerInfo().Addresses) == 0 {
		time.Sleep(time.Millisecond * 250)
	}
	for {
		peerInfo := nd.addressBook.GetLocalPeerInfo()
		closestPeers, err := nd.FindPeersClosestTo(peerInfo.ID, closestPeersToReturn)
		if err != nil {
			logrus.WithError(err).Warnf("refresh could not get peers ids")
			time.Sleep(time.Second * 10)
			continue
		}

		// find peers to announce ourself to
		peerIDs := getPeerIDsFromPeerInfos(closestPeers)

		// announce our peer info to the closest peers
		if err := nd.exchange.Send(ctx, peerInfo.Block(), peerIDs...); err != nil {
			logrus.WithError(err).WithField("peer_ids", peerIDs).Warnf("refresh could not send block")
		}

		// HACK lookup our own peer info just so we can populate our peer table
		nd.GetPeerInfo(ctx, peerInfo.ID)

		// sleep for a bit
		time.Sleep(time.Second * 30)
	}
}

func (nd *DHT) handleBlock(block *blocks.Block) error {
	contentType := block.Type
	switch contentType {
	case PeerInfoRequestType:
		nd.handlePeerInfoRequest(block)
	case peers.PeerInfoType:
		nd.handlePeerInfo(block)
	case ProviderRequestType:
		nd.handleProviderRequest(block)
	case ProviderType:
		nd.handleProvider(block)
	default:
		logrus.WithField("block.PayloadType", contentType).Warn("Payload type not known")
		return nil
	}
	return nil
}

func (nd *DHT) handlePeerInfoRequest(incBlock *blocks.Block) {
	ctx := context.Background()
	payload, ok := incBlock.Payload.(PeerInfoRequest)
	if !ok {
		logrus.Warn("expected PeerInfoRequest, got ", reflect.TypeOf(incBlock.Payload))
		return
	}

	peerInfo, _ := nd.addressBook.GetPeerInfo(payload.PeerID)
	if peerInfo != nil {
		nd.exchange.Send(ctx, peerInfo.Block, incBlock.Metadata.Signer)
		// TODO handle and log error
	}

	closestPeerInfos, _ := nd.FindPeersClosestTo(payload.PeerID, closestPeersToReturn)
	closestBlocks := getBlocksFromPeerInfos(closestPeerInfos)

	for _, block := range closestBlocks {
		if err := nd.exchange.Send(ctx, block, incBlock.Metadata.Signer); err != nil {
			logrus.WithError(err).Warnf("handlePeerInfoRequest could not send block")
			return
		}
	}
}

func (nd *DHT) handlePeerInfo(incBlock *blocks.Block) {
	// TODO handle error
	nd.addressBook.PutPeerInfoFromBlock(incBlock)

	rID := incBlock.GetHeader("requestID")
	if rID == "" {
		return
	}

	q, exists := nd.queries.Load(rID)
	if !exists {
		return
	}

	q.(*query).incomingBlocks <- incBlock
}

func (nd *DHT) handleProviderRequest(incBlock *blocks.Block) {
	ctx := context.Background()
	logger := log.Logger(ctx)
	payload, ok := incBlock.Payload.(ProviderRequest)
	if !ok {
		logger.Warn("expected ProviderRequest", zap.String("actualType", reflect.TypeOf(incBlock.Payload).String()))
		return
	}

	providerBlocks, err := nd.store.GetProviders(payload.Key)
	if err != nil {
		logger.Debug("could not get providers from local store", zap.Error(err))
		// TODO handle and log error
		return
	}

	for _, providerBlock := range providerBlocks {
		cProviderBlock := blocks.Copy(providerBlock)
		cProviderBlock.SetHeader("requestID", payload.RequestID)
		logger.Debug("found provider block", zap.String("blockID", blocks.BestEffortID(providerBlock)))
		nd.exchange.Send(ctx, cProviderBlock, incBlock.Metadata.Signer)
		// TODO handle and log error
	}

	closestPeerInfos, _ := nd.FindPeersClosestTo(payload.Key, closestPeersToReturn)
	closestBlocks := getBlocksFromPeerInfos(closestPeerInfos)

	for _, block := range closestBlocks {
		cBlock := blocks.Copy(block)
		cBlock.SetHeader("requestID", payload.RequestID)
		logger.Debug("sending provider block", zap.String("blockID", blocks.BestEffortID(cBlock)))
		if err := nd.exchange.Send(ctx, cBlock, incBlock.Metadata.Signer); err != nil {
			logger.Warn("handleProviderRequest could not send block", zap.Error(err))
			return
		}
	}
}

func (nd *DHT) handleProvider(incBlock *blocks.Block) {
	ctx := context.Background()
	logger := log.Logger(ctx)

	logger.Debug("handling provider",
		zap.String("blockID", blocks.BestEffortID(incBlock)),
		zap.String("requestID", incBlock.GetHeader("requestID")))

	if err := nd.store.PutProvider(incBlock); err != nil {
		logger.Debug("could not store provider", zap.Error(err))
		// TODO handle error
	}

	rID := incBlock.GetHeader("requestID")
	if rID == "" {
		return
	}

	q, exists := nd.queries.Load(rID)
	if !exists {
		return
	}

	q.(*query).incomingBlocks <- incBlock
}

// FindPeersClosestTo returns an array of n peers closest to the given key by xor distance
func (nd *DHT) FindPeersClosestTo(tk string, n int) ([]*peers.PeerInfo, error) {
	// place to hold the results
	rks := []*peers.PeerInfo{}

	htk := hash(tk)

	peerInfos, _ := nd.addressBook.GetAllPeerInfo()

	// slice to hold the distances
	dists := []distEntry{}
	for _, peerInfo := range peerInfos {
		// calculate distance
		de := distEntry{
			key:      peerInfo.ID,
			dist:     xor([]byte(htk), []byte(hash(peerInfo.ID))),
			peerInfo: peerInfo,
		}
		exists := false
		for _, ee := range dists {
			if ee.key == peerInfo.ID {
				exists = true
				break
			}
		}
		if !exists {
			dists = append(dists, de)
		}
	}

	// sort the distances
	sort.Slice(dists, func(i, j int) bool {
		return lessIntArr(dists[i].dist, dists[j].dist)
	})

	if n > len(dists) {
		n = len(dists)
	}

	// append n the first n number of keys
	for _, de := range dists {
		rks = append(rks, de.peerInfo)
		n--
		if n == 0 {
			break
		}
	}

	return rks, nil
}

// GetPeerInfo returns a peer's info from their id
func (nd *DHT) GetPeerInfo(ctx context.Context, id string) (*peers.PeerInfo, error) {
	q := &query{
		dht:            nd,
		id:             net.RandStringBytesMaskImprSrc(8),
		key:            id,
		queryType:      PeerInfoQuery,
		incomingBlocks: make(chan *blocks.Block),
		outgoingBlocks: make(chan *blocks.Block),
	}

	nd.queries.Store(q.id, q)

	go q.Run(ctx)

	for {
		select {
		case incBlock := <-q.outgoingBlocks:
			// TODO handle error
			nd.addressBook.PutPeerInfoFromBlock(incBlock)
			return nd.addressBook.GetPeerInfo(incBlock.Metadata.Signer)
		case <-time.After(maxQueryTime):
			return nil, ErrNotFound
		case <-ctx.Done():
			return nil, ErrNotFound
		}
	}
}

// PutProviders adds a key of something we provide
// TODO Find a better name for this
func (nd *DHT) PutProviders(ctx context.Context, key string) error {
	block := blocks.NewEphemeralBlock(ProviderType, Provider{
		BlockIDs: []string{key},
	})

	signer := nd.addressBook.GetLocalPeerInfo()
	nd.exchange.Sign(block, signer)

	if err := nd.store.PutProvider(block); err != nil {
		return err
	}

	closestPeers, _ := nd.FindPeersClosestTo(key, closestPeersToReturn)
	closestPeerIDs := getPeerIDsFromPeerInfos(closestPeers)
	if err := nd.exchange.Send(ctx, block, closestPeerIDs...); err != nil {
		logrus.WithError(err).Warnf("PutProviders could not send block")
		return err
	}

	return nil
}

// GetProviders will look for peers that provide a key
func (nd *DHT) GetProviders(ctx context.Context, key string) (chan string, error) {
	q := &query{
		dht:            nd,
		id:             net.RandStringBytesMaskImprSrc(8),
		key:            key,
		queryType:      ProviderQuery,
		incomingBlocks: make(chan *blocks.Block),
		outgoingBlocks: make(chan *blocks.Block),
	}

	nd.queries.Store(q.id, q)

	go q.Run(ctx)

	out := make(chan string, 1)
	go func(q *query, out chan string) {
		defer close(out)
		for {
			select {
			case incBlock := <-q.outgoingBlocks:
				if incBlock == nil {
					// TODO
					log.DefaultLogger.Error("inc block should not be nil",
						zap.Stack("strack"))
				}
				// TODO do we need to check payload and id?
				out <- incBlock.Metadata.Signer
			case <-time.After(maxQueryTime):
				return
			case <-ctx.Done():
				return
			}
		}
	}(q, out)

	return out, nil
}

func (nd *DHT) GetAllProviders() (map[string][]string, error) {
	providers := map[string][]string{}
	blocks, err := nd.store.GetAllProviders()
	if err != nil {
		return nil, err
	}

	for _, block := range blocks {
		payload := block.Payload.(Provider)
		for _, blockID := range payload.BlockIDs {
			if _, ok := providers[blockID]; !ok {
				providers[blockID] = []string{}
			}
			providers[blockID] = append(providers[blockID], block.Metadata.Signer)
		}
	}
	return providers, nil
}

func getPeerIDsFromPeerInfos(peerInfos []*peers.PeerInfo) []string {
	peerIDs := []string{}
	for _, peerInfo := range peerInfos {
		peerIDs = append(peerIDs, peerInfo.ID)
	}
	return peerIDs
}

func getBlocksFromPeerInfos(peerInfos []*peers.PeerInfo) []*blocks.Block {
	blocks := []*blocks.Block{}
	for _, peerInfo := range peerInfos {
		blocks = append(blocks, peerInfo.Block)
	}
	return blocks
}

func blocksOrNil(c []*blocks.Block) []*blocks.Block {
	if len(c) == 0 {
		return nil
	}

	return c
}
