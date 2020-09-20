package tui

import (
	"fmt"
	"github.com/Jahaja/ltt"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type TableSortDirection int

const (
	SortAsc TableSortDirection = iota
	SortDesc
)

type UIView struct {
	sync.Mutex
	Form                 *tview.Form
	Table                *tview.Table
	TableView            string
	TableSortColumnIndex int
	TableSortDirection   TableSortDirection
	Layout               *tview.Flex
	App                  *tview.Application
	StatusText           *tview.TextView
	LoadTest             *ltt.LoadTest
	Client               *LTTClient
	UpdateInterval       int
}

func NewUIView(client *LTTClient, updateInterval int) *UIView {
	ui := &UIView{
		Client:               client,
		UpdateInterval:       updateInterval,
		TableView:            "tasks",
		TableSortColumnIndex: 0,
		TableSortDirection:   SortAsc,
	}
	return ui
}

func (ui *UIView) Setup() {
	lt, err := ui.Client.GetLoadTestInfo()
	if err != nil {
		log.Fatalf("setup: %s\n", err.Error())
	}
	ui.LoadTest = lt

	tview.Styles.PrimitiveBackgroundColor = tcell.Color16
	tview.Styles.BorderColor = tcell.Color231
	tview.Styles.TitleColor = tcell.Color231
	tview.Styles.PrimaryTextColor = tcell.Color231
	tview.Styles.SecondaryTextColor = tcell.Color32

	table := tview.NewTable()
	table.SetTitle("Tasks")
	table.SetBorders(true)
	table.SetBordersColor(tcell.Color231)
	table.SetTitleColor(tcell.Color231)
	table.SetFixed(1, 0)
	table.SetSelectedFunc(func(row, column int) {
		defer ui.Draw()

		if ui.TableSortColumnIndex != column {
			ui.TableSortColumnIndex = column
			return
		}

		if ui.TableSortDirection == SortAsc {
			ui.TableSortDirection = SortDesc
		} else {
			ui.TableSortDirection = SortAsc
		}
	})

	table.SetDoneFunc(func(key tcell.Key) {
		ui.Table.SetSelectable(false, false)
		ui.App.SetFocus(ui.Form)
	})

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'k':
			ui.Table.SetSelectable(false, false)
			ui.App.SetFocus(ui.Form)
		}

		switch event.Key() {
		case tcell.KeyUp:
			ui.Table.SetSelectable(false, false)
			ui.App.SetFocus(ui.Form)
		}

		return event
	})

	form := tview.NewForm()
	form.SetHorizontal(true)
	form.SetButtonBackgroundColor(tcell.ColorDarkSlateGray)
	form.SetButtonTextColor(tcell.Color231)
	form.SetLabelColor(tcell.Color231)
	form.SetFieldTextColor(tcell.Color16)

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		_, btnIdx := ui.Form.GetFocusedItemIndex()
		btnCount := ui.Form.GetButtonCount()
		itemCount := ui.Form.GetFormItemCount()

		switch event.Rune() {
		case 'j':
			ui.Table.SetSelectable(false, true)
			ui.App.SetFocus(ui.Table)
		case 'h':
			ui.App.SetFocus(ui.Table)
			if btnIdx > 0 {
				ui.Form.SetFocus(itemCount + (btnIdx - 1))
			} else {
				ui.Form.SetFocus(0)
			}
			ui.App.SetFocus(ui.Form)
		case 'l':
			if btnIdx < (btnCount - 1) {
				ui.App.SetFocus(ui.Table)
				ui.Form.SetFocus(itemCount + (btnIdx + 1))
				ui.App.SetFocus(ui.Form)
			}
		}

		switch event.Key() {
		case tcell.KeyDown:
			ui.Table.SetSelectable(false, true)
			ui.App.SetFocus(ui.Table)
		case tcell.KeyLeft:
			ui.App.SetFocus(ui.Table)
			if btnIdx > 0 {
				ui.Form.SetFocus(itemCount + (btnIdx - 1))
			} else {
				ui.Form.SetFocus(0)
			}
			ui.App.SetFocus(ui.Form)
		case tcell.KeyRight:
			if btnIdx < (btnCount - 1) {
				ui.App.SetFocus(ui.Table)
				ui.Form.SetFocus(itemCount + (btnIdx + 1))
				ui.App.SetFocus(ui.Form)
			}
		}

		return event
	})

	numUsers := ui.LoadTest.Config.NumUsers
	form.AddButton("Set", func() {
		if err := ui.Client.SetNumUsers(numUsers); err != nil {
			log.Fatal(err.Error())
		}
	})

	confirmStop := tview.NewModal()
	confirmStop.SetText("Are you sure you want to stop?").
		AddButtons([]string{"Stop", "Cancel"}).
		SetDoneFunc(func(index int, label string) {
			if label == "Stop" {
				if err := ui.Client.Stop(); err != nil {
					log.Printf("stop request failed: %s\n", err.Error())
				} else {
					idx := ui.Form.GetButtonIndex("Stop")
					ui.Form.GetButton(idx).SetLabel("Start")
				}
			}

			ui.App.SetRoot(ui.Layout, true).SetFocus(ui.Form)
		}).SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'h', 'l':
			return tcell.NewEventKey(tcell.KeyTab, event.Rune(), tcell.ModNone)
		}

		return event
	})

	onOffBtnLabel := "Start"
	if lt.Status != ltt.StatusStopped {
		onOffBtnLabel = "Stop"
	}

	form.AddButton(onOffBtnLabel, func() {
		btn := ui.Form.GetButton(1)
		if btn.GetLabel() == "Start" {
			if err := ui.Client.Start(); err != nil {
				log.Fatalf("start request failed: %s\n", err.Error())
			} else {
				btn.SetLabel("Stop")
			}
		} else if btn.GetLabel() == "Stop" {
			ui.App.SetRoot(confirmStop, true).SetFocus(confirmStop)
		}
	})

	form.AddButton("Errors", func() {
		btn := ui.Form.GetButton(2)
		if btn.GetLabel() == "Errors" {
			ui.TableView = "errors"
			btn.SetLabel("Tasks")
		} else if btn.GetLabel() == "Tasks" {
			ui.TableView = "tasks"
			btn.SetLabel("Errors")
		}

		ui.Draw()
	})

	confirmReset := tview.NewModal()
	confirmReset.SetText("Are you sure you want to reset?").
		AddButtons([]string{"Reset", "Cancel"}).
		SetDoneFunc(func(index int, label string) {
			if label == "Reset" {
				if err := ui.Client.Reset(); err != nil {
					log.Printf("reset request failed: %s\n", err.Error())
				}
			}

			ui.App.SetRoot(ui.Layout, true).SetFocus(ui.Form)
		}).SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'h', 'l':
			return tcell.NewEventKey(tcell.KeyTab, event.Rune(), tcell.ModNone)
		}

		return event
	})

	form.AddButton("Reset", func() {
		ui.App.SetRoot(confirmReset, true).SetFocus(confirmReset)
	})

	form.AddInputField("Num users:", strconv.Itoa(ui.LoadTest.Config.NumUsers), 8, tview.InputFieldInteger, func(text string) {
		numUsers, _ = strconv.Atoi(text)
	})

	statusText := tview.NewTextView()

	layout := tview.NewFlex()
	layout.SetDirection(tview.FlexRow)
	layout.AddItem(tview.NewTextView().SetText(fmt.Sprintf("Load Testing Tool UI - %s", ui.Client.URI)), 2, 0, false)
	layout.AddItem(statusText, 4, 0, false)
	layout.AddItem(form, 3, 0, true)
	layout.AddItem(table, 0, 10, false)
	app := tview.NewApplication()
	app.SetRoot(layout, true)
	app.EnableMouse(true)
	app.SetFocus(form)

	ui.Form = form
	ui.Layout = layout
	ui.Table = table
	ui.App = app
	ui.StatusText = statusText
}

func (ui *UIView) drawTasksTable() {
	ui.Table.SetCell(0, 0, tview.NewTableCell("Task").SetExpansion(1))
	ui.Table.SetCell(0, 1, tview.NewTableCell(" Num ").SetAlign(tview.AlignCenter))
	ui.Table.SetCell(0, 2, tview.NewTableCell(" Avg ").SetAlign(tview.AlignCenter))
	ui.Table.SetCell(0, 3, tview.NewTableCell(" Median ").SetAlign(tview.AlignCenter))
	ui.Table.SetCell(0, 4, tview.NewTableCell(" 75% ").SetAlign(tview.AlignCenter))
	ui.Table.SetCell(0, 5, tview.NewTableCell(" 85% ").SetAlign(tview.AlignCenter))
	ui.Table.SetCell(0, 6, tview.NewTableCell(" 95% ").SetAlign(tview.AlignCenter))
	ui.Table.SetCell(0, 7, tview.NewTableCell(" 99% ").SetAlign(tview.AlignCenter))
	ui.Table.SetCell(0, 8, tview.NewTableCell(" Num Failed ").SetAlign(tview.AlignCenter))

	sortCell := ui.Table.GetCell(0, ui.TableSortColumnIndex)
	txt := sortCell.Text
	txt = strings.ReplaceAll(txt, " (asc)", "")
	txt = strings.ReplaceAll(txt, " (desc)", "")
	if ui.TableSortDirection == SortAsc {
		sortCell.SetText(fmt.Sprintf("%s (asc)", txt))
	} else {
		sortCell.SetText(fmt.Sprintf("%s (desc)", txt))
	}

	ui.Table.SetCell(0, ui.TableSortColumnIndex, sortCell)

	tasks := make([]*ltt.TaskStats, 0, len(ui.LoadTest.Stats.Tasks))
	for _, t := range ui.LoadTest.Stats.Tasks {
		tasks = append(tasks, t)
	}

	sort.Slice(tasks, func(i, j int) bool {
		switch ui.TableSortColumnIndex {
		case 0:
			if ui.TableSortDirection == SortDesc {
				return strings.Compare(tasks[i].Name, tasks[j].Name) == 1
			} else {
				return strings.Compare(tasks[i].Name, tasks[j].Name) == -1
			}
		case 1:
			if ui.TableSortDirection == SortDesc {
				return tasks[i].TotalRuns > tasks[j].TotalRuns
			} else {
				return tasks[i].TotalRuns < tasks[j].TotalRuns
			}
		case 2:
			if ui.TableSortDirection == SortDesc {
				return tasks[i].AverageDuration > tasks[j].AverageDuration
			} else {
				return tasks[i].AverageDuration < tasks[j].AverageDuration
			}
		case 3:
			if ui.TableSortDirection == SortDesc {
				return tasks[i].Percentiles[50] > tasks[j].Percentiles[50]
			} else {
				return tasks[i].Percentiles[50] < tasks[j].Percentiles[50]
			}
		case 4:
			if ui.TableSortDirection == SortDesc {
				return tasks[i].Percentiles[75] > tasks[j].Percentiles[75]
			} else {
				return tasks[i].Percentiles[75] < tasks[j].Percentiles[75]
			}
		case 5:
			if ui.TableSortDirection == SortDesc {
				return tasks[i].Percentiles[85] > tasks[j].Percentiles[85]
			} else {
				return tasks[i].Percentiles[85] < tasks[j].Percentiles[85]
			}
		case 6:
			if ui.TableSortDirection == SortDesc {
				return tasks[i].Percentiles[95] > tasks[j].Percentiles[95]
			} else {
				return tasks[i].Percentiles[95] < tasks[j].Percentiles[95]
			}
		case 7:
			if ui.TableSortDirection == SortDesc {
				return tasks[i].Percentiles[99] > tasks[j].Percentiles[99]
			} else {
				return tasks[i].Percentiles[99] < tasks[j].Percentiles[99]
			}
		case 8:
			if ui.TableSortDirection == SortDesc {
				return tasks[i].NumFailed > tasks[j].NumFailed
			} else {
				return tasks[i].NumFailed < tasks[j].NumFailed
			}
		}

		return strings.Compare(tasks[i].Name, tasks[j].Name) == -1
	})

	i := 1
	for _, t := range tasks {
		ui.Table.SetCell(i, 0, tview.NewTableCell(t.Name).SetExpansion(1))
		ui.Table.SetCell(i, 1, tview.NewTableCell(fmt.Sprintf(" %s ", strconv.FormatInt(t.TotalRuns, 10))).SetAlign(tview.AlignCenter))
		ui.Table.SetCell(i, 2, tview.NewTableCell(fmt.Sprintf(" %.1f ", t.AverageDuration)).SetAlign(tview.AlignCenter))

		j := 3
		for _, v := range t.Percentiles {
			ui.Table.SetCell(i, j, tview.NewTableCell(strconv.FormatInt(v, 10)).SetAlign(tview.AlignCenter))
			j++
		}

		ui.Table.SetCell(i, j, tview.NewTableCell(strconv.FormatInt(t.NumFailed, 10)).SetAlign(tview.AlignCenter))
		i++
	}
}

func (ui *UIView) drawErrorsTable() {
	ui.Table.SetCell(0, 0, tview.NewTableCell("Task"))
	ui.Table.SetCell(0, 1, tview.NewTableCell("Message").SetExpansion(1))
	ui.Table.SetCell(0, 2, tview.NewTableCell("Num").SetAlign(tview.AlignCenter))

	tasks := make([]*ltt.TaskStats, 0, len(ui.LoadTest.Stats.Tasks))
	for _, t := range ui.LoadTest.Stats.Tasks {
		tasks = append(tasks, t)
	}

	sort.Slice(tasks, func(i, j int) bool {
		return strings.Compare(tasks[i].Name, tasks[j].Name) == -1
	})

	i := 1
	for _, t := range tasks {
		for msg, count := range t.Errors {
			ui.Table.SetCell(i, 0, tview.NewTableCell(t.Name))
			ui.Table.SetCell(i, 1, tview.NewTableCell(msg).SetExpansion(1))
			ui.Table.SetCell(i, 2, tview.NewTableCell(strconv.FormatInt(count, 10)).SetAlign(tview.AlignCenter))
			i++
		}
	}
}

func (ui *UIView) Draw() {
	ui.Lock()
	defer ui.Unlock()

	if ui.LoadTest == nil {
		return
	}

	stats := ui.LoadTest.Stats

	var failRate float64
	if stats.NumTotal == 0 {
		failRate = 0
	} else {
		failRate = (float64(stats.NumFailed) / float64(stats.NumTotal)) * 100
	}

	ui.StatusText.SetText(fmt.Sprintf(" %d tasks running since %s\n Total: %d, Successful: %d, Failed: %d, Fail rate: %.1f%%\n Users: %d, RPS: %.1f, Avg duration: %.1f ms\n Status: %s",
		len(stats.Tasks), stats.StartTime.Format("2006-01-02 15:04:05"), stats.NumTotal, stats.NumSuccessful,
		stats.NumFailed, failRate, stats.NumUsers, stats.CurrentRPS, stats.AverageDuration, ui.LoadTest.Status))

	ui.Table.Clear()
	if ui.TableView == "tasks" {
		ui.drawTasksTable()
	} else if ui.TableView == "errors" {
		ui.drawErrorsTable()
	}
}

func (ui *UIView) Run() {
	ui.Setup()

	go func() {
		for {
			lt, err := ui.Client.GetLoadTestInfo()
			if err != nil {
				log.Printf("periodical update: %w", err.Error())
			}
			ui.LoadTest = lt

			ui.App.QueueUpdateDraw(func() {
				ui.Draw()
			})

			time.Sleep(time.Second * time.Duration(ui.UpdateInterval))
		}
	}()

	if err := ui.App.Run(); err != nil {
		panic(err)
	}
}
