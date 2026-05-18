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
