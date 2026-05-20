package tmux

import (
	"fmt"
	"os/exec"

	"github.com/stefanschmerda/tmux-commander/internal/theme"
)

func Installed() bool {
	_, err := exec.LookPath("tmux")
	return err == nil
}

func PopupCommand(binary string, width string, height string) string {
	return fmt.Sprintf("tmux display-popup -E -w %s -h %s %s", width, height, binary)
}

func PopupArgs(binary string, width string, height string, border bool, t theme.Theme) []string {
	args := []string{"display-popup", "-E"}
	if !border {
		args = append(args, "-B")
	}
	if style := PopupStyle(t); style != "" {
		args = append(args, "-s", style)
	}
	if borderStyle := PopupBorderStyle(t); borderStyle != "" {
		args = append(args, "-S", borderStyle)
	}
	if width != "" {
		args = append(args, "-w", width)
	}
	if height != "" {
		args = append(args, "-h", height)
	}
	return append(args, binary)
}

func OpenPopup(binary string, width string, height string, border bool, t theme.Theme) error {
	cmd := exec.Command("tmux", PopupArgs(binary, width, height, border, t)...)
	return cmd.Run()
}

func PopupStyle(t theme.Theme) string {
	return style(t.Query, t.Background)
}

func PopupBorderStyle(t theme.Theme) string {
	return style(t.CommanderBorder, t.Background)
}

func style(fg string, bg string) string {
	if fg == "" && bg == "" {
		return ""
	}
	if fg == "" {
		return "bg=" + bg
	}
	if bg == "" {
		return "fg=" + fg
	}
	return "fg=" + fg + ",bg=" + bg
}
