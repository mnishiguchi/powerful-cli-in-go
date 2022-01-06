package app

import (
	"context"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/widgets/donut"
	"github.com/mum4k/termdash/widgets/segmentdisplay"
	"github.com/mum4k/termdash/widgets/text"
)

type widgets struct {
	// widgets
	timerDonutWidget *donut.Donut                   // the donut widget in the timer section
	displayType      *segmentdisplay.SegmentDisplay // the text widget in the timer section
	infoTextWidget   *text.Text                     // the text widget in the info section
	timerTextWidget  *text.Text                     // the text widget in the timer section
	// GO channels to update widget concurrently
	timerDonutUpdaterCh chan []int
	infoTextUpdaterCh   chan string
	timerTextUpdaterCh  chan string
	textTypeUpdaterCh   chan string
}

// Update updates the widgets with new data.
func (w *widgets) update(timerValues []int, textType, infoText, timerText string, redrawerCh chan<- bool) {
	if infoText != "" {
		w.infoTextUpdaterCh <- infoText
	}

	if textType != "" {
		w.textTypeUpdaterCh <- textType
	}

	if timerText != "" {
		w.timerTextUpdaterCh <- timerText
	}

	if len(timerValues) > 0 {
		w.timerDonutUpdaterCh <- timerValues
	}

	redrawerCh <- true
}

func newWidgets(ctx context.Context, errCh chan<- error) (*widgets, error) {
	w := &widgets{}
	var err error

	w.timerDonutUpdaterCh = make(chan []int)
	w.textTypeUpdaterCh = make(chan string)
	w.infoTextUpdaterCh = make(chan string)
	w.timerTextUpdaterCh = make(chan string)

	w.timerDonutWidget, err = newDonut(ctx, w.timerDonutUpdaterCh, errCh)
	if err != nil {
		return nil, err
	}

	w.displayType, err = newSegmentDisplay(ctx, w.textTypeUpdaterCh, errCh)
	if err != nil {
		return nil, err
	}

	w.infoTextWidget, err = newText(ctx, w.infoTextUpdaterCh, errCh)
	if err != nil {
		return nil, err
	}

	w.timerTextWidget, err = newText(ctx, w.timerTextUpdaterCh, errCh)
	if err != nil {
		return nil, err
	}

	return w, nil
}

func newText(ctx context.Context, textUpdaterCh <-chan string, errCh chan<- error) (*text.Text, error) {
	txt, err := text.New()
	if err != nil {
		return nil, err
	}

	// Goroutine to update Text
	go func() {
		for {
			select {
			case t := <-textUpdaterCh:
				txt.Reset()
				errCh <- txt.Write(t)

			case <-ctx.Done():
				return
			}
		}
	}()

	return txt, nil
}

func newDonut(ctx context.Context, donutUpdaterCh <-chan []int, errCh chan<- error) (*donut.Donut, error) {
	don, err := donut.New(
		donut.Clockwise(),
		donut.CellOpts(cell.FgColor(cell.ColorBlue)),
	)

	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case d := <-donutUpdaterCh:
				if d[0] <= d[1] {
					errCh <- don.Absolute(d[0], d[1])
				}

			case <-ctx.Done():
				return
			}
		}
	}()

	return don, nil
}

func newSegmentDisplay(ctx context.Context, textUpdaterCh <-chan string, errCh chan<- error) (*segmentdisplay.SegmentDisplay, error) {
	sd, err := segmentdisplay.New()
	if err != nil {
		return nil, err
	}

	// Goroutine to update SegmentDisplay
	go func() {
		for {
			select {
			case t := <-textUpdaterCh:
				if t == "" {
					t = " "
				}

				errCh <- sd.Write([]*segmentdisplay.TextChunk{
					segmentdisplay.NewChunk(t),
				})
			case <-ctx.Done():
				return
			}
		}
	}()

	return sd, nil
}
