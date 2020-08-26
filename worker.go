package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/govice/golinks/blockmap"
	"github.com/spf13/viper"
)

type Worker struct {
	daemon *daemon
}

func NewWorker(daemon *daemon) (*Worker, error) {
	return &Worker{daemon: daemon}, nil
}

var ErrBadRootPath = errors.New("bad root_path")

func (w *Worker) Execute(ctx context.Context) error {
	logln("starting worker")
	// TODO pull this from its own config file(s)
	rootPath := viper.GetString("root_path")
	fi, err := os.Stat(rootPath)
	if err != nil || !fi.IsDir() {
		logln(err)
		return ErrBadRootPath
	}
	absRootPath, err := filepath.Abs(rootPath)
	if err != nil {
		return err
	}
	period := viper.GetInt("generation_period")
	logln("generation_period:", period, "ms")
	generationTicker := time.NewTicker(time.Duration(period) * time.Millisecond)
	logln("generating startup blockmap")
	if err := w.generateBlockmap(absRootPath); err != nil {
		errln("initial blockmap generation failed")
		return err
	}

	logln("scheduling periodic blockmap generation")
	for {
		select {
		case <-ctx.Done():
			logln("received termination on worker context")
			generationTicker.Stop()
			return nil //TODO err canceled?
		case <-generationTicker.C:
			logln("generating scheduled blockmap for tick")
			if err := w.generateBlockmap(absRootPath); err != nil {
				errln("scheduled blockmap generation failed")
				return err
			}
		}
	}
}

func (w *Worker) generateBlockmap(rootPath string) error {
	logln("initializing blockmap for", rootPath)
	blkmap := blockmap.New(rootPath)
	if err := blkmap.Generate(); err != nil {
		errln("failed to generate blockmap for", rootPath, err)
		return err
	}

	jobsDir := filepath.Join(w.daemon.HomeDir(), "jobs")
	//TODO handle error
	os.Mkdir(jobsDir, os.ModePerm)

	fileUUID, err := uuid.NewRandom()
	if err != nil {
		errln("failed to generate blockmap uuid", err)
		return err
	}
	if err := blkmap.SaveNamed(jobsDir, fileUUID.String()); err != nil {
		errln("failed to save blockmap", err)
		return err
	}

	logln("saved blockmap", fileUUID.String())

	return nil
}
