package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"

	"github.com/golang/protobuf/proto"
	"golang.org/x/net/http2"
)

const maxMsgSize = 1024 * 1024 * 1024

func dumpMsg(dataBuf *bytes.Buffer, net string, path string, frame *http2.DataFrame, side int) {
	buf := frame.Data()
	if len(buf) == 0 {
		return
	}

	id := frame.Header().StreamID
	compress := buf[0]
	length := int(binary.BigEndian.Uint32(buf[1:5]))
	if compress == 1 {
		// use compression, check Message-Encoding later
		log.Printf("%s %d use compression, msg %q", net, id, buf[5:])
		return
	}

	dataBuf.Write(buf[5:])
	// We may capture the middle frame of a message, so the length is not right here
	// So here we do a assumption, if the length or dataBuf length is larger than 1GB,
	// we think we can't decode this message
	if length > maxMsgSize || dataBuf.Len() > maxMsgSize {
		dataBuf.Truncate(0)
		return
	}

	if length != dataBuf.Len() {
		return
	}
	data := dataBuf.Bytes()
	defer func() {
		dataBuf.Truncate(0)
	}()

	// Em, a little ugly here, try refactor later.
	if msgs, ok := pathMsgs[path]; ok {
		switch side {
		case 1:
			msg := proto.Clone(msgs[0])
			if err := proto.Unmarshal(data, msg); err == nil {
				log.Printf("%s %d %d %s %s", net, id, length, path, msg)
				return
			}
		case 2:
			msg := proto.Clone(msgs[1])
			if err := proto.Unmarshal(data, msg); err == nil {
				log.Printf("%s %d %d %s %s", net, id, length, path, msg)
				return
			}
		default:
			// We can't know the data is request or response
			for _, msg := range msgs {
				msg := proto.Clone(msg)
				if err := proto.Unmarshal(data, msg); err == nil {
					log.Printf("%s %d %d %s %s", net, id, length, path, msg)
					return
				}
			}
		}
	}

	dumpProto(net, id, path, data)
}

func dumpProto(net string, id uint32, path string, buf []byte) {
	var out bytes.Buffer
	if err := decodeProto(&out, buf, 0); err != nil {
		// decode failed
		log.Printf("%s %d %d %s %q", net, id, len(buf), path, buf)
	} else {
		log.Printf("%s %d %d %s\n%s", net, id, len(buf), path, out.String())
	}
}

func decodeProto(out *bytes.Buffer, buf []byte, depth int) error {
out:
	for {
		if len(buf) == 0 {
			return nil
		}

		for i := 0; i < depth; i++ {
			out.WriteString("  ")
		}

		op, n := proto.DecodeVarint(buf)
		if n == 0 {
			return io.ErrUnexpectedEOF
		}

		buf = buf[n:]

		tag := op >> 3
		wire := op & 7

		switch wire {
		default:
			fmt.Fprintf(out, "tag=%d unknown wire=%d\n", tag, wire)
			break out
		case proto.WireBytes:
			l, n := proto.DecodeVarint(buf)
			if n == 0 {
				return io.ErrUnexpectedEOF
			}
			buf = buf[n:]
			if len(buf) < int(l) {
				return io.ErrUnexpectedEOF
			}

			// Here we can't know the raw bytes is string, or embedded message
			// So we try to parse like a embedded message at first
			outLen := out.Len()
			fmt.Fprintf(out, "tag=%d struct\n", tag)
			if err := decodeProto(out, buf[0:int(l)], depth+1); err != nil {
				// Seem this is not a embedded message, print raw buffer
				out.Truncate(outLen)
				fmt.Fprintf(out, "tag=%d bytes=%q\n", tag, buf[0:int(l)])
			}
			buf = buf[l:]
		case proto.WireFixed32:
			if len(buf) < 4 {
				return io.ErrUnexpectedEOF
			}
			u := binary.LittleEndian.Uint32(buf[0:4])
			buf = buf[4:]
			fmt.Fprintf(out, "tag=%d fix32=%d\n", tag, u)
		case proto.WireFixed64:
			if len(buf) < 8 {
				return io.ErrUnexpectedEOF
			}
			u := binary.LittleEndian.Uint64(buf[0:8])
			buf = buf[4:]
			fmt.Fprintf(out, "tag=%d fix64=%d\n", tag, u)
		case proto.WireVarint:
			u, n := proto.DecodeVarint(buf)
			if n == 0 {
				return io.ErrUnexpectedEOF
			}
			buf = buf[n:]
			fmt.Fprintf(out, "tag=%d varint=%d\n", tag, u)
		case proto.WireStartGroup:
			fmt.Fprintf(out, "tag=%d start\n", tag)
			depth++
		case proto.WireEndGroup:
			fmt.Fprintf(out, "tag=%d end\n", tag)
			depth--
		}
	}
	return io.ErrUnexpectedEOF
}
