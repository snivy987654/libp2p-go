package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
	"strings"
	"sync"
	"net/http"
	"path/filepath"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/multiformats/go-multiaddr"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/pnet"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/core/network"
	tcp "github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/libp2p/go-libp2p/p2p/muxer/yamux"
	"io"
	"crypto/sha256"
	"github.com/ipfs/go-cid"
)

// IPCMessage represents the structure for MsgPack communication with Python
type IPCMessage struct {
	Type      string                 `msgpack:"type"`
	Topic     string                 `msgpack:"topic,omitempty"`
	Data      map[string]interface{} `msgpack:"data"`
	NodeID    string                 `msgpack:"node_id,omitempty"`
	Timestamp float64                `msgpack:"timestamp"`
}

const (
	SocketPath = "/dev/shm/.sys_ipc.sock"
	TopicCmd    = "ds_global_commands"
	TopicScan   = "ds_scan_results"
)

var (
	topics = make(map[string]*pubsub.Topic)
	topicsMu sync.Mutex
	
	// Map to store local file paths indexed by a simple key/hash
	localFiles = make(map[string]string)
	filesMu    sync.RWMutex
	
	httpBridgePort int
)

const BitswapProtocol = protocol.ID("/ds/bitswap/1.0.0")

func getTopic(ps *pubsub.PubSub, name string) (*pubsub.Topic, error) {
	topicsMu.Lock()
	defer topicsMu.Unlock()
	if t, ok := topics[name]; ok {
		return t, nil
	}
	t, err := ps.Join(name)
	if err != nil {
		return nil, err
	}
	topics[name] = t
	return t, nil
}

func startHttpBridge(ctx context.Context, h host.Host, dht *dht.IpfsDHT) {
	mux := http.NewServeMux()
	mux.HandleFunc("/bitswap/", func(w http.ResponseWriter, r *http.Request) {
		// ... logic remains same ...
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) < 2 {
			http.Error(w, "invalid path", 400)
			return
		}
		key := parts[1]

		filesMu.RLock()
		filePath, ok := localFiles[key]
		filesMu.RUnlock()
		
		if ok {
			f, err := os.Open(filePath)
			if err != nil {
				http.Error(w, "error opening file", 500)
				return
			}
			defer f.Close()
			w.Header().Set("Content-Type", "application/octet-stream")
			_, _ = io.Copy(w, f)
			return
		}
		http.Error(w, "file not found locally", 404)
	})

	// Port Stepping Implementation
	for port := 8081; port <= 8090; port++ {
		addr := fmt.Sprintf(":%d", port)
		srv := &http.Server{Addr: addr, Handler: mux}
		l, err := net.Listen("tcp", addr)
		if err == nil {
			httpBridgePort = port
			fmt.Printf("🌐 HTTP Bridge active on %s\n", addr)
			go srv.Serve(l)
			return
		}
		log.Printf("⚠️ Port %d occupied, stepping...", port)
	}
	log.Printf("❌ Failed to bind HTTP Bridge to any port in range 8081-8090")
}

func setupBitswapProtocol(h host.Host) {
	h.SetStreamHandler(BitswapProtocol, func(s network.Stream) {
		defer s.Close()
		
		// Read requested file key (simplified)
		buf := make([]byte, 256)
		n, err := s.Read(buf)
		if err != nil { return }
		key := string(buf[:n])

		filesMu.RLock()
		path, ok := localFiles[key]
		filesMu.RUnlock()

		if ok {
			f, err := os.Open(path)
			if err == nil {
				defer f.Close()
				_, _ = io.Copy(s, f)
			}
		}
	})
}

func main() {
	// Flags for sandbox isolation
	swarmKeyPath := flag.String("pnetkey", "", "Path to the 256-bit swarm key for private network isolation")
	identityKeyPath := flag.String("key", "", "Path to the identity private key file")
	writeIDPath := flag.String("write-id", "", "Path to write the Peer ID to")
	ipcPath := flag.String("ipc", "/dev/shm/.sys_ipc.sock", "Path for the Unix Domain Socket")
	bootstrapNodes := flag.String("bootstrap", "", "Comma-separated list of bootstrap multiaddresses")
	listenAddr := flag.String("listen", "/ip4/0.0.0.0/tcp/4001", "Listening multiaddress")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var opts []libp2p.Option
	opts = append(opts, 
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.Security(noise.ID, noise.New),
		libp2p.Muxer("/yamux/1.0.0", yamux.DefaultTransport),
		libp2p.ListenAddrStrings(*listenAddr),
	)

	// 0. Identity Key Implementation
	if *identityKeyPath != "" {
		var priv crypto.PrivKey
		if _, err := os.Stat(*identityKeyPath); os.IsNotExist(err) {
			// Create new key
			priv, _, err = crypto.GenerateKeyPair(crypto.Ed25519, -1)
			if err != nil {
				log.Fatalf("Failed to generate key: %v", err)
			}
			keyBytes, _ := crypto.MarshalPrivateKey(priv)
			os.WriteFile(*identityKeyPath, keyBytes, 0600)
			fmt.Printf("🔑 Generated new identity key: %s\n", *identityKeyPath)
		} else {
			// Load existing key
			keyBytes, err := os.ReadFile(*identityKeyPath)
			if err != nil {
				log.Fatalf("Failed to read key: %v", err)
			}
			priv, err = crypto.UnmarshalPrivateKey(keyBytes)
			if err != nil {
				log.Fatalf("Failed to unmarshal key: %v", err)
			}
			fmt.Printf("🔑 Loaded identity key: %s\n", *identityKeyPath)
		}
		opts = append(opts, libp2p.Identity(priv))
	}

	// 1. Private Network (pnet) Implementation
	if *swarmKeyPath != "" {
		keyFile, err := os.Open(*swarmKeyPath)
		if err != nil {
			log.Fatalf("Failed to open swarm key: %v", err)
		}
		psk, err := pnet.DecodeV1PSK(keyFile)
		if err != nil {
			log.Fatalf("Failed to decode swarm key: %v", err)
		}
		opts = append(opts, libp2p.PrivateNetwork(psk))
		fmt.Printf("🔒 Private Network active (Swarm Key: %s)\n", *swarmKeyPath)
	}

	// Initialize Libp2p Host
	h, err := libp2p.New(opts...)
	if err != nil {
		log.Fatalf("Failed to start host: %v", err)
	}
	defer h.Close()

	fmt.Printf("🚀 ds_core_network active | ID: %s\n", h.ID().String())
	fmt.Printf("📡 Listening on: %v\n", h.Addrs())

	// Write Peer ID to file if requested
	if *writeIDPath != "" {
		err := os.WriteFile(*writeIDPath, []byte(h.ID().String()), 0644)
		if err != nil {
			log.Printf("Failed to write Peer ID to file: %v", err)
		} else {
			fmt.Printf("📝 Peer ID written to %s\n", *writeIDPath)
		}
	}

	// 2. Initialize Kademlia DHT
	kDHT, err := dht.New(ctx, h, dht.Mode(dht.ModeServer))
	if err != nil {
		log.Fatalf("Failed to start DHT: %v", err)
	}

	// 3. Connect to Bootstrap Nodes
	if *bootstrapNodes != "" {
		for _, addrStr := range strings.Split(*bootstrapNodes, ",") {
			addr, err := multiaddr.NewMultiaddr(addrStr)
			if err != nil {
				log.Printf("Invalid bootstrap address: %s", addrStr)
				continue
			}
			pinfo, err := peer.AddrInfoFromP2pAddr(addr)
			if err != nil {
				log.Printf("Failed to get peer info: %v", err)
				continue
			}
			if err := h.Connect(ctx, *pinfo); err != nil {
				log.Printf("Failed to connect to bootstrap node %s: %v", pinfo.ID, err)
			} else {
				fmt.Printf("🔗 Connected to bootstrap node: %s\n", pinfo.ID)
			}
		}
	}

	if err = kDHT.Bootstrap(ctx); err != nil {
		log.Printf("DHT Bootstrap warning: %v", err)
	}

	// 4. Initialize GossipSub
	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		log.Fatalf("Failed to start GossipSub: %v", err)
	}

	// 4.1 Join Core Topics immediately to ensure GossipSub mesh forms
	cmdTopic, err := getTopic(ps, TopicCmd)
	if err == nil {
		_, _ = cmdTopic.Subscribe() 
	}
	scanTopic, err := getTopic(ps, TopicScan)
	if err == nil {
		_, _ = scanTopic.Subscribe()
	}

	// 5. Setup Bitswap Protocol Handler
	setupBitswapProtocol(h)

	// 6. Start IPC Server (UDS + MsgPack)
	go startIPCServer(ctx, ps, h, kDHT, *ipcPath)

	// 7. Start HTTP Bridge for Scavenger scripts
	startHttpBridge(ctx, h, kDHT)

	// Wait for termination
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	fmt.Println("Shutting down...")
}

func startIPCServer(ctx context.Context, ps *pubsub.PubSub, h host.Host, dht *dht.IpfsDHT, path string) {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("Failed to create IPC directory: %v", err)
	}

	_ = os.Remove(path)
	l, err := net.Listen("unix", path)
	if err != nil {
		log.Fatalf("Failed to listen on IPC socket %s: %v", path, err)
	}
	fmt.Printf("🔌 IPC Server active on %s\n", path)
	defer l.Close()
	_ = os.Chmod(path, 0666)

	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}
		go handleIPCClient(ctx, conn, ps, h, dht)
	}
}

func handleIPCClient(ctx context.Context, conn net.Conn, ps *pubsub.PubSub, h host.Host, dht *dht.IpfsDHT) {
	defer conn.Close()
	dec := msgpack.NewDecoder(conn)
	enc := msgpack.NewEncoder(conn)

	// Forward messages from core topics to this IPC client
	if ps == nil {
		log.Printf("❌ GossipSub NOT initialized")
		return
	}
	cmdTopic, err := getTopic(ps, TopicCmd)
	if err != nil {
		log.Printf("❌ Failed to get topic %s: %v", TopicCmd, err)
		return
	}
	cmdSub, err := cmdTopic.Subscribe()
	if err != nil {
		log.Printf("❌ Failed to subscribe to topic %s: %v", TopicCmd, err)
		return
	}
	defer cmdSub.Cancel()
	go forwardMessagesToPython(ctx, cmdSub, enc)

	for {
		var msg IPCMessage
		if err := dec.Decode(&msg); err != nil {
			return
		}

		switch msg.Type {
		case "broadcast":
			topic, err := getTopic(ps, msg.Topic)
			if err != nil {
				log.Printf("❌ Broadcast fail: %v", err)
				continue
			}
			data, _ := msgpack.Marshal(msg.Data)
			_ = topic.Publish(ctx, data)
		case "provide_file":
			filePath := msg.Data["path"].(string)
			name := msg.Data["name"].(string) // Use 'name' as our CID-equivalent key
			
			filesMu.Lock()
			localFiles[name] = filePath
			filesMu.Unlock()
			
			// Announce to DHT
			hsh := sha256.Sum256([]byte(name))
			c := cid.NewCidV1(cid.Raw, hsh[:])
			if err := dht.Provide(ctx, c, true); err != nil {
				log.Printf("DHT Provide error: %v", err)
			}
			
			fmt.Printf("📦 Providing file: %s as key: %s\n", filePath, name)
			_ = enc.Encode(map[string]string{"status": "providing", "cid": name})

		case "fetch_file":
			key := msg.Data["cid"].(string)
			outPath := msg.Data["path"].(string)
			
			fmt.Printf("📥 Fetching file key: %s...\n", key)
			
			// 1. Find providers in DHT
			hsh := sha256.Sum256([]byte(key))
			c := cid.NewCidV1(cid.Raw, hsh[:])
			providers, err := dht.FindProviders(ctx, c)
			if err != nil || len(providers) == 0 {
				_ = enc.Encode(map[string]string{"error": "no providers found"})
				continue
			}

			// 2. Connect and pull
			success := false
			for _, p := range providers {
				if p.ID == h.ID() { continue }
				s, err := h.NewStream(ctx, p.ID, BitswapProtocol)
				if err != nil { continue }
				
				_, _ = s.Write([]byte(key))
				
				outF, err := os.Create(outPath)
				if err != nil { s.Close(); break }
				
				_, err = io.Copy(outF, s)
				outF.Close()
				s.Close()
				
				if err == nil {
					success = true
					break
				}
			}

			if success {
				fmt.Printf("✅ Fetched file key: %s to %s\n", key, outPath)
				_ = enc.Encode(map[string]string{"status": "fetched", "path": outPath})
			} else {
				_ = enc.Encode(map[string]string{"error": "fetch failed"})
			}
		case "get_status":
			_ = enc.Encode(map[string]interface{}{
				"status": "online",
				"peers":  len(h.Network().Peers()),
				"id":      h.ID().String(),
				"addrs":   h.Addrs(),
				"http_bridge_port": httpBridgePort,
				"pid":     os.Getpid(),
			})
		}
	}
}

func forwardMessagesToPython(ctx context.Context, sub *pubsub.Subscription, enc *msgpack.Encoder) {
	for {
		msg, err := sub.Next(ctx)
		if err != nil {
			return
		}
		var data map[string]interface{}
		_ = msgpack.Unmarshal(msg.Data, &data)

		out := IPCMessage{
			Type:      "mesh_event",
			Topic:     sub.Topic(),
			Data:      data,
			NodeID:    msg.ReceivedFrom.String(),
			Timestamp: float64(time.Now().Unix()),
		}
		_ = enc.Encode(out)
	}
}
