package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func chooseProfile(profilesPath string, private bool) {
	profiles := listProfiles(profilesPath, private)
	if len(profiles) == 0 {
		fmt.Println("No profiles found")
		return
	}

	var items []string
	for _, p := range profiles {
		items = append(items, p.Name)
	}

	m := profileModel{
		choices:  items,
		profiles: profiles,
		private:  private,
		path:     profilesPath,
	}

	p := tea.NewProgram(m, tea.WithInput(os.Stderr), tea.WithOutput(os.Stderr))
	finalModel, err := p.Run()
	if err != nil {
		return
	}

	if finalModel.(profileModel).cursor < len(profiles) {
		selected := profiles[finalModel.(profileModel).cursor]
		
		if finalModel.(profileModel).view {
			// View file content
			if !selected.IsDir {
				viewProfileFile(selected.Path)
				// Show menu again after viewing
				chooseProfile(profilesPath, private)
			}
			return
		}
		
		if finalModel.(profileModel).selected {
			if selected.IsDir {
				chooseProfile(selected.Path, private)
			} else {
				loadProfileFile(selected.Path)
			}
		}
	}
}

type profileModel struct {
	choices  []string
	profiles []Profile
	cursor   int
	selected bool
	view     bool
	private  bool
	path     string
}

func (m profileModel) Init() tea.Cmd {
	return nil
}

func (m profileModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "v":
			m.view = true
			return m, tea.Quit
		case "enter", " ":
			m.selected = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m profileModel) View() string {
	s := "Select Profile:\n\n"
	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}
	s += "\nPress v to view, Enter to select, q to quit.\n"
	return s
}

func setNamespace(namespace string) {
	cmd := exec.Command("kubectl", "config", "set-context", "--current", fmt.Sprintf("--namespace=%s", namespace))
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error setting namespace: %v\n%s\n", err, output)
		return
	}
	fmt.Fprintf(os.Stderr, "Namespace set to: %s\n", namespace)
}

func selectNamespace() {
	cmd := exec.Command("kubectl", "get", "namespaces", "-o", "jsonpath={.items[*].metadata.name}")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error getting namespaces: %v\n", err)
		return
	}

	namespaces := strings.Fields(string(output))
	if len(namespaces) == 0 {
		fmt.Println("No namespaces found")
		return
	}

	m := namespaceModel{
		choices: namespaces,
	}

	p := tea.NewProgram(m, tea.WithInput(os.Stderr), tea.WithOutput(os.Stderr))
	finalModel, err := p.Run()
	if err != nil {
		return
	}

	if finalModel.(namespaceModel).selected && finalModel.(namespaceModel).cursor < len(namespaces) {
		selectedNamespace := namespaces[finalModel.(namespaceModel).cursor]
		setNamespace(selectedNamespace)
	}
}

type namespaceModel struct {
	choices  []string
	cursor   int
	selected bool
}

func (m namespaceModel) Init() tea.Cmd {
	return nil
}

func (m namespaceModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			m.selected = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m namespaceModel) View() string {
	s := "Select Namespace:\n\n"
	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}
	s += "\nPress q to quit.\n"
	return s
}

func setAWSProfile(profile string) {
	fmt.Printf("export AWS_PROFILE=%s\n", profile)
	fmt.Printf("export AWS_REGION=$(aws configure get region --profile %s)\n", profile)
	fmt.Println("unset AWS_DEFAULT_REGION KUBE_CONFIG_PATH KUBECONFIG")
}

func selectAWSProfile() {
	fmt.Println("aws configure list-profiles")
}

func viewProfileFile(path string) {
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}
	fmt.Println("\n" + string(content))
	fmt.Println("\nPress Enter to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}
