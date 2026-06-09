package pkg

import (
	"bytes"
	"encoding/hex"
	"log/slog"
	"strings"
)

func ParseI2CDump(dump string) []byte {
	slog.Debug("======== I2C Dump ========")
	slog.Debug("\n" + dump)
	slog.Debug("======== ======== ========")

	lines := strings.Split(dump, "\n")

	buf := make([]byte, 0, 1024)
	w := bytes.NewBuffer(buf)

	for _, line := range lines[1:17] {

		p1 := strings.Split(line, ": ")[1]
		p2 := strings.Split(p1, "    ")[0]

		for _, x := range strings.Split(p2, " ") {
			b, err := hex.DecodeString(x)
			if err != nil {
				slog.Error("could not parse", "code", err)
				panic(err)
			}
			w.Write(b)
		}

	}

	allBytes := w.Bytes()
	slog.Debug("raw_decoded_i2c_bytes", slog.String("hex_string", hex.EncodeToString(allBytes)))
	return allBytes

}
