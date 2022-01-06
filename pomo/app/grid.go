package app

import (
	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/container/grid"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/terminalapi"
)

func newGrid(b *buttonSet, w *widgets,
	t terminalapi.Terminal) (*container.Container, error) {

	builder := grid.New()

	// Add first row
	builder.Add(
		grid.RowHeightPerc(30,
			grid.ColWidthPercWithOpts(30,
				[]container.Option{
					container.Border(linestyle.Light),
					container.BorderTitle("Press Q to Quit"),
				},
				// Add inside row
				grid.RowHeightPerc(80,
					grid.Widget(w.timerDonutWidget)),
				grid.RowHeightPercWithOpts(20,
					[]container.Option{
						container.AlignHorizontal(align.HorizontalCenter),
					},
					grid.Widget(w.timerTextWidget,
						container.AlignHorizontal(align.HorizontalCenter),
						container.AlignVertical(align.VerticalMiddle),
						container.PaddingLeftPercent(49),
					),
				),
			),
			grid.ColWidthPerc(70,
				grid.RowHeightPerc(80,
					grid.Widget(w.displayType, container.Border(linestyle.Light)),
				),
				grid.RowHeightPerc(20,
					grid.Widget(w.infoTextWidget, container.Border(linestyle.Light)),
				),
			),
		),
	)

	// Add second row
	builder.Add(
		grid.RowHeightPerc(10,
			grid.ColWidthPerc(50,
				grid.Widget(b.btStart),
			),
			grid.ColWidthPerc(50,
				grid.Widget(b.btPause),
			),
		),
	)

	// Add third row
	builder.Add(
		grid.RowHeightPerc(60),
	)

	gridOpts, err := builder.Build()
	if err != nil {
		return nil, err
	}

	c, err := container.New(t, gridOpts...)
	if err != nil {
		return nil, err
	}

	return c, nil
}
