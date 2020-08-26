package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/govice/golinks/block"
	"github.com/govice/golinks/blockchain"
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
	if err := w.generateAndUploadBlockmap(absRootPath); err != nil {
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
			if err := w.generateAndUploadBlockmap(absRootPath); err != nil {
				errln("scheduled blockmap generation failed")
				return err
			}
		}
	}
}

func (w *Worker) generateAndUploadBlockmap(rootPath string) error {
	w.daemon.chainMutex.Lock()
	defer w.daemon.chainMutex.Unlock()
	blkmap, err := w.generateBlockmap(rootPath)
	if err != nil {
		errln("failed to generate blockmap", err)
		return err
	}

	blockmapBytes, err := json.Marshal(blkmap)
	if err != nil {
		errln("failed to marshal blockmap")
		return err
	}

	localHeadBlock, err := w.daemon.chainTracker.LocalHead()
	if err != nil {
		errln("failed to get local head block", err)
		return err
	}

	stagedBlock := block.NewSHA512(localHeadBlock.Index+1, blockmapBytes, localHeadBlock.BlockHash)

	subchain := &blockchain.Blockchain{
		Blocks: []block.Block{*localHeadBlock, *stagedBlock},
	}

	if err := subchain.Validate(); err != nil {
		errln("failed to validate subchain")
		return err
	}

	if err := w.uploadBlock(stagedBlock); err != nil {
		errln("failed to upload staged block")
		return err
	}

	return nil
}

func (w *Worker) generateBlockmap(rootPath string) (*blockmap.BlockMap, error) {
	logln("generating blockmap for", rootPath)
	blkmap := blockmap.New(rootPath)
	if err := blkmap.Generate(); err != nil {
		errln("failed to generate blockmap for", rootPath, err)
		return nil, err
	}

	return blkmap, nil
}

func (w *Worker) uploadBlock(blk *block.Block) error {
	return w.daemon.golinksService.UploadBlock(blk)
}
