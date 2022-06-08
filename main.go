package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"
	. "github.com/lxn/walk/declarative"

	"github.com/micmonay/keybd_event"
)

type sMainWindow struct {
	*walk.MainWindow
}

type IP struct {
	// Index int
	Addr string
	// Baz   float64
	// Quux    time.Time
	checked bool
}

type IPModel struct {
	walk.TableModelBase
	walk.SorterBase
	sortColumn int
	sortOrder  walk.SortOrder
	items      []*IP
}

func NewIPModel() *IPModel {
	m := new(IPModel)
	//m.ResetRows()
	return m
}

// Called by the TableView from SetModel and every time the model publishes a
// RowsReset event.
func (m *IPModel) RowCount() int {
	return len(m.items)
}

// Called by the TableView when it needs the text to display for a given cell.
func (m *IPModel) Value(row, col int) interface{} {
	item := m.items[row]

	switch col {
	case 0:
		return item.Addr
	}

	panic("unexpected col")
}

// Called by the TableView to retrieve if a given row is checked.
func (m *IPModel) Checked(row int) bool {
	return m.items[row].checked
}

// Called by the TableView when the user toggled the check box of a given row.
func (m *IPModel) SetChecked(row int, checked bool) error {
	m.items[row].checked = checked

	return nil
}

// Called by the TableView to sort the model.
func (m *IPModel) Sort(col int, order walk.SortOrder) error {
	m.sortColumn, m.sortOrder = col, order

	sort.SliceStable(m.items, func(i, j int) bool {
		a, b := m.items[i], m.items[j]

		c := func(ls bool) bool {
			if m.sortOrder == walk.SortAscending {
				return ls
			}

			return !ls
		}

		switch m.sortColumn {
		case 0:
			return c(a.Addr < b.Addr)
		}

		panic("unreachable")
	})

	return m.SorterBase.Sort(col, order)
}

func (m *IPModel) ResetRows() {
	// Create some random data.
	m.items = make([]*IP, rand.Intn(100))

	// Notify TableView and other interested parties about the reset.
	m.PublishRowsReset()

	m.Sort(m.sortColumn, m.sortOrder)
}

func (m *IPModel) AddRow(keybd keybd_event.KeyBonding) {
	// ipAddr, err := ipify.GetIp()
	// if err != nil {
	// 	fmt.Println("err : ", err)
	// } else {
	// 	fmt.Println("ip : ", ipAddr)
	// }

	url := "https://api64.ipify.org"
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	ip, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	//fmt.Println(string(ip))
	ipAddr := string(ip)
	for _, item := range m.items {
		if strings.Contains(item.Addr, ipAddr) {
			fmt.Println("Same IP Detect")
			keybd.Launching()
			return
		}
	}

	m.items = append(m.items, &IP{Addr: ipAddr})
	m.PublishRowsReset()
}

func (m *IPModel) ClearRows() {
	for _, item := range m.items {
		if item.checked {
			fmt.Println("checked")
		}
	}

	m.items = m.items[:0]
	m.PublishRowsReset()
}

func (m *IPModel) RemoveRow() {
	for i, item := range m.items {
		if item.checked {
			m.items = append(m.items[:i], m.items[i+1:]...)
		}
	}
	m.PublishRowsReset()
}

func main() {
	rand.Seed(time.Now().UnixNano())

	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		panic(err)
	}
	kb.SetKeys(keybd_event.VK_ESC)

	//mw := new(sMainWindow)
	model := NewIPModel()
	var tv *walk.TableView

	MainWindow{
		Title:  "IP Checker",
		Size:   declarative.Size{Width: 120, Height: 240},
		Layout: VBox{MarginsZero: true},
		Children: []Widget{
			TableView{
				AssignTo:         &tv,
				AlternatingRowBG: true,
				CheckBoxes:       true,
				ColumnsOrderable: true,
				MultiSelection:   true,
				Columns: []TableViewColumn{
					{Title: "IP", Width: 118},
				},

				Model: model,
				OnSelectedIndexesChanged: func() {
					fmt.Printf("SelectedIndexes: %v\n", tv.SelectedIndexes())
				},
			},
			PushButton{
				Text: "추가",
				OnClicked: func() {
					model.AddRow(kb)
				},
			},
			PushButton{
				Text:      "초기화",
				OnClicked: model.ClearRows,
			},
			// PushButton{
			// 	Text:      "선택 삭제",
			// 	OnClicked: model.RemoveRow,
			// },
		},
		Bounds: Rectangle{X: 1289, Y: 570},
	}.Run()
}
