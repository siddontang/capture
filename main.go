// Copyright 2012 Google, Inc. All rights reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file in the root of the source
// tree.

// This binary provides sample code for using the gopacket TCP assembler and TCP
// stream reader.  It reads packets off the wire and reconstructs HTTP requests
// it sees, logging them.
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/hpack"

	"github.com/google/gopacket"
	"github.com/google/gopacket/examples/util"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"
)

var iface = flag.String("i", "eth0", "Interface to get packets from")
var fname = flag.String("r", "", "Filename to read from, overrides -i")
var snaplen = flag.Int("s", 65536, "SnapLen for pcap packet capture")
var filter = flag.String("f", "tcp and dst port 80", "BPF filter for pcap")
var logAllPackets = flag.Bool("v", false, "Logs every packet in great detail")
var pprofAddr = flag.String("a", "127.0.0.1:6060", "Pprof address")

// Build a simple HTTP request parser using tcpassembly.StreamFactory and tcpassembly.Stream interfaces

// httpStreamFactory implements tcpassembly.StreamFactory
type httpStreamFactory struct{}

// httpStream will handle the actual decoding of http requests.
type httpStream struct {
	net, transport gopacket.Flow
	r              tcpreader.ReaderStream
}

func (h *httpStreamFactory) New(net, transport gopacket.Flow) tcpassembly.Stream {
	hstream := &httpStream{
		net:       net,
		transport: transport,
		r:         tcpreader.NewReaderStream(),
	}
	go hstream.run() // Important... we must guarantee that data from the reader stream is read.

	// ReaderStream implements tcpassembly.Stream, so we can return a pointer to it.
	return &hstream.r
}

const connPreface string = "PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n"

func parseRequest(net string, prefix string, buf *bufio.Reader) (bool, error) {
	prefix = strings.ToUpper(prefix)
	if strings.HasPrefix(prefix, "GET") || strings.HasPrefix(prefix, "POST") ||
		strings.HasPrefix(prefix, "PUT") || strings.HasPrefix(prefix, "DELET") ||
		strings.HasPrefix(prefix, "HEAD") {
		r, err := http.ReadRequest(buf)
		if err != nil {
			return false, err
		}

		log.Printf("%s Req %v", net, r)

		buf.Discard(int(r.ContentLength))
		r.Body.Close()
		return true, nil
	}
	return false, nil
}

func parseResponse(net string, prefix string, buf *bufio.Reader) (bool, error) {
	prefix = strings.ToUpper(prefix)
	if strings.HasPrefix(prefix, "HTTP") {
		resp, err := http.ReadResponse(buf, nil)
		if err != nil {
			return false, err
		}

		log.Printf("%s Resp %v", net, resp)

		buf.Discard(int(resp.ContentLength))
		resp.Body.Close()
		return true, nil
	}
	return false, nil
}

var readNum int64

var streamPath = map[string]map[uint32]string{}
var pathLock sync.RWMutex

func checkSlow(reason string, net string) chan struct{}{
	ch := make(chan struct{})
	go func() {
	t := time.NewTicker(30 * time.Second)
	defer t.Stop()

	for {
		select {
		case <-ch:
			return 
		case <-t.C:
			log.Printf("failed to %s data at %s", reason, net)
		}
	}
	}()
	return ch
}

func (h *httpStream) run() {
	buf := bufio.NewReader(&h.r)
	framer := http2.NewFramer(ioutil.Discard, buf)
	framer.MaxHeaderListSize = uint32(16 << 20)
	framer.ReadMetaHeaders = hpack.NewDecoder(4096, nil)
	net := fmt.Sprintf("%s:%s -> %s:%s", h.net.Src(), h.transport.Src(), h.net.Dst(), h.transport.Dst())
	revNet := fmt.Sprintf("%s:%s -> %s:%s", h.net.Dst(), h.transport.Dst(), h.net.Src(), h.transport.Src())
	// 1 request, 2 response, 0 unkonwn
	var streamSide = map[uint32]int{}
	var dataBufs = map[uint32]bytes.Buffer{}

	defer func() {
		pathLock.Lock()
		delete(streamPath, net)
		delete(streamPath, revNet)
		pathLock.Unlock()
	}()
	var ch chan struct{}
	for {
		ch = checkSlow("peek", net)
		peekBuf, err := buf.Peek(9)
		close(ch)
		if err == io.EOF {
			return
		} else if err != nil {
			log.Print("Error reading frame", h.net, h.transport, ":", err)
			continue
		}

		prefix := string(peekBuf)

		// log.Printf("%s prefix %q", net, prefix)
		if ok, err := parseRequest(net, prefix, buf); ok || err != nil {
			continue
		}

		if ok, err := parseResponse(net, prefix, buf); ok || err != nil {
			continue
		}

		if strings.HasPrefix(prefix, "PRI") {
			buf.Discard(len(connPreface))
		}

		ch = checkSlow("read", net)
		frame, err := framer.ReadFrame()
		close(ch)
		if err == io.EOF {
			return
		}

		if err != nil {
			log.Print("Error reading frame", h.net, h.transport, ":", err)
			continue
		}

		id := frame.Header().StreamID
		// log.Printf("%s id %d frame %v", net, id, frame.Header())
		switch frame := frame.(type) {
		case *http2.MetaHeadersFrame:
			for _, hf := range frame.Fields {
				// log.Printf("%s id %d %s=%s", net, id, hf.Name, hf.Value)
				if hf.Name == ":path" {
					// TODO: remove stale stream ID
					pathLock.Lock()
					_, ok := streamPath[net]
					if !ok {
						streamPath[net] = map[uint32]string{}
					}
					streamPath[net][id] = hf.Value
					pathLock.Unlock()
					streamSide[id] = 1
				} else if hf.Name == ":status" {
					streamSide[id] = 2
				}
			}
		case *http2.DataFrame:
			var path string
			pathLock.RLock()
			nets, ok := streamPath[net]
			if !ok {
				nets, ok = streamPath[revNet]
			}

			if ok {
				path = nets[id]
			}

			pathLock.RUnlock()
			// log.Printf("%s id %d path %s, data", net, id, path)
			dataBuf, ok := dataBufs[id]
			if !ok {
				dataBuf = bytes.Buffer{}
				dataBufs[id] = dataBuf
			}
			dumpMsg(&dataBuf, net, path, frame, streamSide[id])
		default:
		}
	}
}

func main() {
	defer util.Run()()
	go func() {
		http.ListenAndServe(*pprofAddr, nil)
	}()

	var handle *pcap.Handle
	var err error

	// Set up pcap packet capture
	if *fname != "" {
		log.Printf("Reading from pcap dump %q", *fname)
		handle, err = pcap.OpenOffline(*fname)
	} else {
		log.Printf("Starting capture on interface %q", *iface)
		handle, err = pcap.OpenLive(*iface, int32(*snaplen), true, pcap.BlockForever)
	}
	if err != nil {
		log.Fatal(err)
	}

	if err := handle.SetBPFFilter(*filter); err != nil {
		log.Fatal(err)
	}

	// Set up assembly
	streamFactory := &httpStreamFactory{}
	streamPool := tcpassembly.NewStreamPool(streamFactory)
	assembler := tcpassembly.NewAssembler(streamPool)

	log.Println("reading in packets")
	// Read in packets, pass to assembler.
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	packets := packetSource.Packets()
	ticker := time.Tick(time.Second * 10)
	for {
		select {
		case packet := <-packets:
			// A nil packet indicates the end of a pcap file.
			if packet == nil {
				return
			}
			if *logAllPackets {
				log.Println(packet)
			}
			if packet.NetworkLayer() == nil || packet.TransportLayer() == nil || packet.TransportLayer().LayerType() != layers.LayerTypeTCP {
				log.Println("Unusable packet")
				continue
			}
			tcp := packet.TransportLayer().(*layers.TCP)
			assembler.AssembleWithTimestamp(packet.NetworkLayer().NetworkFlow(), tcp, packet.Metadata().Timestamp)

		case <-ticker:
			// Every 10s, flush connections that haven't seen activity in the past 20s.
			assembler.FlushOlderThan(time.Now().Add(time.Second * -20))
		}
	}
}
