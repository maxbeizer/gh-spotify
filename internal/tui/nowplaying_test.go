package tui

import (
	"fmt"
	"image"
	"image/color"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/maxbeizer/gh-spotify/internal/art"
	"github.com/maxbeizer/gh-spotify/internal/spotify"
)

func playingModel(t *testing.T) NowPlayingModel {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 4), G: 200, B: uint8(y * 4), A: 255})
		}
	}
	m := NewNowPlayingModel(&spotify.Client{}, DefaultTheme())
	m.state = &spotify.PlayerState{
		ProgressMS: 112_000,
		Item: &spotify.Track{
			ID:       "t1",
			Name:     "Herz und Mund und Tat und Leben, Cantata BWV 147: Jesu, Joy of Man's Desiring (Arr. Myra Hess for Piano)",
			Artists:  []spotify.Artist{{Name: "Johann Sebastian Bach"}, {Name: "Rafał Blechacz"}},
			Album:    spotify.Album{Name: "Johann Sebastian Bach"},
			Duration: 243_000,
		},
	}
	m.artCache = art.RenderFromImage(img, artWidth)
	m.artTrackID = "t1"
	return m
}

func TestNowPlayingViewFitsAnySize(t *testing.T) {
	m := playingModel(t)
	for _, w := range []int{10, 30, 50, 80, 200} {
		for h := 1; h <= 40; h++ {
			t.Run(fmt.Sprintf("%dx%d", w, h), func(t *testing.T) {
				out := m.View(w, h)
				if got := lipgloss.Height(out); got > h {
					t.Errorf("height %d exceeds %d:\n%s", got, h, out)
				}
				if got := lipgloss.Width(out); got > w {
					t.Errorf("width %d exceeds %d:\n%s", got, w, out)
				}
			})
		}
	}
}

func TestNowPlayingViewStaysCompact(t *testing.T) {
	m := playingModel(t)
	out := m.View(120, 60)
	if got := lipgloss.Height(out); got > 7 {
		t.Errorf("card should stay small on tall terminals, got %d lines:\n%s", got, out)
	}
	if !strings.Contains(out, "█") || !strings.Contains(out, "╭") {
		t.Errorf("expected bordered card with art:\n%s", out)
	}
}

func TestAppNowPlayingAnchoredToBottom(t *testing.T) {
	a := App{width: 100, height: 40, theme: DefaultTheme(), nowPlaying: playingModel(t)}
	a.controls = NewControlsModel(&spotify.Client{}, a.theme)
	lines := strings.Split(a.renderPanel(), "\n")
	if strings.TrimSpace(lines[0]) != "" {
		t.Errorf("expected empty space above the player, got first line %q", lines[0])
	}
	if strings.TrimSpace(lines[len(lines)-1]) == "" {
		t.Errorf("expected controls on the last line of the panel")
	}
}
