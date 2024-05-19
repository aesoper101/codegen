package commands

import (
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var _ tea.Model = (*installEnvModel)(nil)

type nextInstallStepMsg struct {
	path string
}

type finishInstallMsg struct{}

type execErrMsg struct {
	path string
	err  error
}

var checkEnvList = []string{
	"kitex",
	"hz",
	"thriftgo",
}

var defaultInstallPath = map[string]string{
	"thriftgo": "github.com/cloudwego/thriftgo@latest",
	"kitex":    "github.com/cloudwego/kitex/tool/cmd/kitex@latest",
	"hz":       "github.com/cloudwego/hertz/cmd/hz@latest",
}

func envCheckPreRunE(_ *cobra.Command, _ []string) error {
	_, err := tea.NewProgram(newInstallEnvModel()).Run()
	return err
}

func initCheckInfo() tea.Cmd {
	for _, env := range checkEnvList {
		if _, err := exec.LookPath(env); err != nil {
			return func() tea.Msg {
				return nextInstallStepMsg{
					path: env,
				}
			}
		}
	}
	return func() tea.Msg {
		return nil
	}
}

func nextInstall(path string) tea.Cmd {
	c := exec.Command("go", "install", defaultInstallPath[path])
	return tea.ExecProcess(c, func(err error) tea.Msg {
		if err != nil {
			return execErrMsg{path: path, err: err}
		}

		for _, env := range checkEnvList {
			if env == path {
				continue
			}
			if _, err := exec.LookPath(env); err != nil {
				return nextInstallStepMsg{path: env}
			}
		}
		return nil
	})
}

type installEnvModel struct {
	currentInstall string
	err            error
}

func newInstallEnvModel() *installEnvModel {
	return &installEnvModel{}
}

func (m installEnvModel) Init() tea.Cmd {
	return initCheckInfo()
}

func (m installEnvModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}

	case nextInstallStepMsg:
		m.currentInstall = msg.path
		return m, nextInstall(msg.path)
	case execErrMsg:
		m.err = msg.err
		return m, tea.Quit
	case finishInstallMsg:
		return m, tea.Quit
	}

	return m, tea.Quit
}

func (m installEnvModel) View() string {
	if m.err != nil {
		return m.err.Error() + "\n please press ctrl+c to quit"
	}

	if m.currentInstall != "" {
		return "installing " + m.currentInstall + "\n please wait or press ctrl+c to quit"
	}
	return ""
}
