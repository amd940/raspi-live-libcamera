package main

import (
	"errors"

	"github.com/amd940/raspi-live-libcamera/internal/ffmpeg/dash"
	"github.com/amd940/raspi-live-libcamera/internal/libcameravid"
	"github.com/amd940/raspi-live-libcamera/internal/server"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// DashCfg represents the DASH configuration options
type DashCfg struct {
	Video        *VideoCfg
	Port         int
	Directory    string
	TLSCert      string
	TLSKey       string
	SegmentTime  int    // Segment length target duration in seconds
	PlaylistSize int    // Maximum number of playlist entries
	StorageSize  int    // Maximum number of unreferenced segments to keep on disk before removal
	CORS         string // https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS
}

func newDashCmd(video *VideoCfg) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dash",
		Short: "Stream video using DASH",
		Long:  "Stream video using DASH",
	}

	cfg := DashCfg{
		Video: video,
	}

	cmd.Flags().IntVar(&cfg.Port, "port", 0, "static file server port")

	cmd.Flags().StringVar(&cfg.Directory, "directory", "", "static file server directory")

	cmd.Flags().StringVar(&cfg.TLSCert, "tls-cert", "", "static file server TLS certificate")

	cmd.Flags().StringVar(&cfg.TLSKey, "tls-key", "", "static file server TLS key")

	cmd.Flags().IntVar(&cfg.SegmentTime, "segment-time", 2, "target segment duration in seconds")

	cmd.Flags().IntVar(&cfg.PlaylistSize, "playlist-size", 10, "maximum number of playlist entries")

	cmd.Flags().IntVar(&cfg.StorageSize, "storage-size", 1, "maximum number of unreferenced segments to keep on disk before removal")

	cmd.Flags().StringVar(&cfg.CORS, "cors", "", "whether or not to include a CORS header")

	cmd.Flags().SortFlags = false

	cmd.Run = func(cmd *cobra.Command, args []string) {
		streamDash(cfg)
	}

	return cmd
}

func streamDash(cfg DashCfg) {
	// Set up libcameravid stream
	raspiOptions := libcameravid.Options{
		Width:          cfg.Video.Width,
		Height:         cfg.Video.Height,
		Fps:            cfg.Video.Fps,
		HorizontalFlip: cfg.Video.HorizontalFlip,
		VerticalFlip:   cfg.Video.VerticalFlip,
	}
	raspiStream, err := libcameravid.NewStream(raspiOptions)
	if err != nil {
		log.Fatal().Msg("Encountered an error streaming video from the Raspberry Pi Camera Module")
	}

	// Set up DASH muxer
	muxer := dash.Muxer{
		Directory: cfg.Directory,
		Options: dash.Options{
			Fps:          cfg.Video.Fps,
			SegmentTime:  cfg.SegmentTime,
			PlaylistSize: cfg.PlaylistSize,
			StorageSize:  cfg.StorageSize,
		},
	}

	// Set up static file server
	srv := server.Static{
		Port:      cfg.Port,
		Directory: cfg.Directory,
		Cert:      cfg.TLSCert,
		Key:       cfg.TLSKey,
		CORS:      cfg.CORS,
	}

	// Set up a channel for exiting
	stop := make(chan struct{})
	osStopper(stop)

	// Serve files generated by the video stream
	go func() {
		err := srv.ListenAndServe()
		if errors.Is(err, server.ErrInvalidDirectory) {
			log.Fatal().Msg("Directory does not exist")
		}
		if err != nil {
			log.Debug().Err(err).Msg("Encountered an error serving video")
			log.Fatal().Msg("Encountered an error serving video")
		}
		stop <- struct{}{}
	}()

	// Stream video
	go func() {
		if err := muxDash(raspiStream, &muxer); err != nil {
			log.Fatal().Msg("Encountered an error muxing video")
		}
		stop <- struct{}{}
	}()

	// Wait for a stop signal
	<-stop

	log.Info().Msg("Shutting down")

	raspiStream.Video.Close()
	srv.Shutdown(serverShutdownDeadline)
}

func muxDash(raspiStream *libcameravid.Stream, muxer *dash.Muxer) error {
	if err := muxer.Mux(raspiStream.Video); err != nil {
		log.Debug().Err(err).Msg("Encountered an error starting video mux")
		return err
	}
	log.Debug().Str("cmd", muxer.String()).Msg("Started ffmpeg muxer")

	if err := raspiStream.Start(); err != nil {
		log.Debug().Err(err).Msg("Encountered an error starting video stream")
		return err
	}
	log.Debug().Str("cmd", raspiStream.String()).Msg("Started libcamera-vid")

	if err := muxer.Wait(); err != nil {
		log.Debug().Err(err).Msg("Encountered an error waiting for video mux")
		return err
	}

	if err := raspiStream.Wait(); err != nil {
		log.Debug().Err(err).Msg("Encountered an error waiting for video stream")
		return err
	}

	return nil
}
