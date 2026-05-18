package tmux

import (
	"fmt"
	"os/exec"
)

func Installed() bool {
	_, err := exec.LookPath("tmux")
	return err == nil
}

func PopupCommand(binary string, width string, height string) string {
	return fmt.Sprintf("tmux display-popup -E -w %s -h %s %s", width, height, binary)
}

func PopupArgs(binary string, width string, height string, border bool) []string {
	args := []string{"display-popup", "-E"}
	if !border {
		args = append(args, "-B")
	}
	if width != "" {
		args = append(args, "-w", width)
	}
	if height != "" {
		args = append(args, "-h", height)
	}
	return append(args, binary)
}

func OpenPopup(binary string, width string, height string, border bool) error {
	cmd := exec.Command("tmux", PopupArgs(binary, width, height, border)...)
	return cmd.Run()
}
