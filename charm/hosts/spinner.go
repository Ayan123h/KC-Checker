package hosts

import (
	"KC-Checker/common"
	"KC-Checker/helper"
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"os"
	"sort"
	"strings"
	"time"
)

type errMsg error

type model struct {
	spinner  spinner.Model
	quitting bool
	err      error
}

var data = make(map[string]string)
var finished = false

func initialModel() model {
	s := spinner.New()
	s.Spinner = spinner.Meter
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#624CAB"))
	return model{spinner: s}
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	checked := true
	for _, f := range data {
		if f == "?" {
			checked = false
		}
	}

	if checked {
		finished = true
		go helper.Dispatcher(helper.ToProxies(helper.GetInput("proxies.txt")))
		time.Sleep(time.Second * 2)
		m.View()
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			finished = true
			m.View()
			return m, tea.Quit
		default:
			return m, nil
		}

	case errMsg:
		m.err = msg
		return m, nil

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m model) View() string {
	if m.err != nil {
		return m.err.Error()
	}

	if finished {
		return ""
	}

	if m.quitting {
		return ""
	}
	str := fmt.Sprintf("\n\n   %s Loading hosts, please wait\n\n", m.spinner.View())

	re := lipgloss.NewRenderer(os.Stdout)
	baseStyle := re.NewStyle().Padding(0, 1)
	headerStyle := baseStyle.Copy().Foreground(lipgloss.Color("252")).Bold(true)
	selectedStyle := baseStyle.Copy().Foreground(lipgloss.Color("#01BE85")).Background(lipgloss.Color("#00432F"))
	warningStyle := baseStyle.Copy().Foreground(lipgloss.Color("#BEAA01")).Background(lipgloss.Color("#414300"))
	errorStyle := baseStyle.Copy().Foreground(lipgloss.Color("#BE0101")).Background(lipgloss.Color("#430000"))

	headers := []string{"Name", "Time"}

	if len(data) == 0 {
		for _, val := range common.GetConfig().Hosts {
			data[val] = "?"
		}
	}

	hosts := common.CurrentCheckedHosts

	for _, val := range hosts {
		response := val.ResponseTime.String()

		if response == "999h0m0s" {
			response = "error"
		}

		data[val.Host] = response
	}

	var dataString [][]string

	var keys []string
	for key := range data {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		dataString = append(dataString, []string{key, data[key]})
	}

	CapitalizeHeaders := func(data []string) []string {
		for i := range data {
			data[i] = strings.ToUpper(data[i])
		}
		return data
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(re.NewStyle().Foreground(lipgloss.Color("238"))).
		Headers(CapitalizeHeaders(headers)...).
		Width(80).
		Rows(dataString...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return headerStyle
			}

			if len(common.GetHosts()) > 0 && dataString[row-1][0] == common.GetHosts()[0].Host {
				return selectedStyle
			}

			if dataString[row-1][1] == "error" {
				return errorStyle
			}

			if dataString[row-1][1] == "?" {
				return warningStyle
			}

			if row%2 == 0 {
				return baseStyle.Copy().Foreground(lipgloss.Color("245"))
			}
			return baseStyle.Copy().Foreground(lipgloss.Color("252"))
		})

	return str + t.Render() + "\n"
}

func Run() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
