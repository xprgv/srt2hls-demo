package main

import (
	_ "embed"
	"log"
	"net/http"
	"time"

	hls "github.com/bluenviron/gohlslib"
	"github.com/bluenviron/gohlslib/pkg/codecs"
	"github.com/bluenviron/mediacommon/pkg/formats/mpegts"
	srt "github.com/datarhei/gosrt"
	"github.com/spf13/pflag"
)

var (
	srtInput    = pflag.String("input", "localhost:3000", "")
	httpAddress = pflag.String("http", "localhost:8080", "")
)

//go:embed index.html
var index []byte

func main() {
	pflag.Parse()

	hlsMuxer := hls.Muxer{
		VideoTrack: &hls.Track{
			Codec: &codecs.H264{},
		},
		// AudioTrack: &hls.Track{
		// 	Codec: &codecs.Opus{},
		// },

		Variant: hls.MuxerVariantLowLatency,
	}

	if err := hlsMuxer.Start(); err != nil {
		log.Fatal(err)
	}

	httpServer := &http.Server{
		Addr:    *httpAddress,
		Handler: handleIndex(hlsMuxer.Handle),
	}

	go func() {
		log.Println("Starting http on:", *httpAddress)

		if err := httpServer.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	srtConfig := srt.DefaultConfig()
	conn, err := srt.Dial("srt", *srtInput, srtConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = conn.Close() }()

	log.Println("SRT connected to:", *srtInput)

	mpegtsReader, err := mpegts.NewReader(mpegts.NewBufferedReader(conn))
	if err != nil {
		log.Fatal(err)
	}

	var timeDecoder *mpegts.TimeDecoder

	found := false
	for _, track := range mpegtsReader.Tracks() {
		if _, ok := track.Codec.(*mpegts.CodecH264); ok {
			mpegtsReader.OnDataH264(track, func(rawPTS, dts int64, au [][]byte) error {
				if timeDecoder == nil {
					timeDecoder = mpegts.NewTimeDecoder(rawPTS)
				}
				pts := timeDecoder.Decode(rawPTS)

				return hlsMuxer.WriteH264(time.Now(), pts, au)
			})

			found = true
			break
		}
	}

	if !found {
		log.Fatal("streams not found")
	}

	for {
		if err := mpegtsReader.Read(); err != nil {
			log.Fatal("Failed to read mpegts packet from srt data stream:", err.Error())
		}
	}
}

func handleIndex(wrapped http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(index))
			return
		}

		wrapped(w, r)
	}
}
