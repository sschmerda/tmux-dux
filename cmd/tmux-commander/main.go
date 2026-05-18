package main

import (
	"fmt"
	"os"

	"github.com/stefanschmerda/tmux-commander/internal/actions"
	"github.com/stefanschmerda/tmux-commander/internal/config"
	"github.com/stefanschmerda/tmux-commander/internal/palette"
)

func main() {
	cfg, _, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "tmux-commander: load config: %v\n", err)
		os.Exit(1)
	}

	selected, err := palette.Run(cfg.Commands)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tmux-commander: run palette: %v\n", err)
		os.Exit(1)
	}
	if selected == nil {
		return
	}

	action, err := actions.Build(*selected, cfg.UI)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tmux-commander: build action: %v\n", err)
		os.Exit(1)
	}
	if err := actions.Dispatch(action); err != nil {
		fmt.Fprintf(os.Stderr, "tmux-commander: dispatch action: %v\n", err)
		os.Exit(1)
	}
}
