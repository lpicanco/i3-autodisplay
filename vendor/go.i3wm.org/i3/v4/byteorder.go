package i3

import (
	"encoding/binary"
	"io"
	"strings"
)

// detectByteOrder sends messages to i3 to determine the byte order it uses.
// For details on this technique, see:
// https://build.i3wm.org/docs/ipc.html#_appendix_a_detecting_byte_order_in_memory_safe_languages
func detectByteOrder(conn io.ReadWriter) (binary.ByteOrder, error) {
	const (
		// targetLen is 0x00 01 01 00 in both, big and little endian
		targetLen = 65536 + 256

		// SUBSCRIBE was introduced in 3.e (2010-03-30)
		prefixSubscribe = "[]"

		// RUN_COMMAND was always present
		prefixCmd = "nop byte-order detection. padding: "
	)

	// 2. Send a big endian encoded message of type SUBSCRIBE:
	payload := []byte(prefixSubscribe + strings.Repeat(" ", targetLen-len(prefixSubscribe)))
	if err := binary.Write(conn, binary.BigEndian, &header{magic, uint32(len(payload)), messageTypeSubscribe}); err != nil {
		return nil, err
	}
	if _, err := conn.Write(payload); err != nil {
		return nil, err
	}

	// 3. Send a byte order independent RUN_COMMAND message:
	payload = []byte(prefixCmd + strings.Repeat("a", targetLen-len(prefixCmd)))
	if err := binary.Write(conn, binary.BigEndian, &header{magic, uint32(len(payload)), messageTypeRunCommand}); err != nil {
		return nil, err
	}
	if _, err := conn.Write(payload); err != nil {
		return nil, err
	}

	// 4. Receive a message header, decode the message type as big endian:
	var header [14]byte
	if _, err := io.ReadFull(conn, header[:]); err != nil {
		return nil, err
	}
	if messageType(binary.BigEndian.Uint32(header[10:14])) == messageReplyTypeCommand {
		order := binary.LittleEndian // our big endian message was not answered
		// Read remaining payload
		_, err := io.ReadFull(conn, make([]byte, order.Uint32(header[6:10])))
		return order, err
	}
	order := binary.BigEndian // our big endian message was answered
	// Read remaining payload
	if _, err := io.ReadFull(conn, make([]byte, order.Uint32(header[6:10]))); err != nil {
		return order, err
	}

	// Slurp the pending RUN_COMMAND reply.
	sock := &socket{conn: conn, order: order}
	_, err := sock.recvMsg()
	return binary.BigEndian, err
}
