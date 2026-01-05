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
	chooseProfileWithSearch(profilesPath, private, "")
}

func chooseProfileWithSearch(profilesPath string, private bool, initialSearch string) {
	profiles := listProfiles(profilesPath, private)
	if len(profiles) == 0 {
		fmt.Fprintln(os.Stderr, "No profiles found")
		return
	}

	var items []string
	for _, p := range profiles {
		items = append(items, p.Name)
	}

	m := profileModel{
		choices:     items,
		profiles:    profiles,
		private:     private,
		path:        profilesPath,
		searchMode:  initialSearch != "",
		searchInput: initialSearch,
	}

	if initialSearch != "" {
		m.filteredItems = m.filterProfiles()
	}

	p := tea.NewProgram(m, tea.WithInput(os.Stderr), tea.WithOutput(os.Stderr))
	finalModel, err := p.Run()
	if err != nil {
		return
	}

	final := finalModel.(profileModel)

	// Determine the actual profile index
	var profileIndex int
	var validSelection bool

	if final.searchMode && len(final.filteredItems) > 0 {
		if final.cursor < len(final.filteredItems) {
			profileIndex = final.filteredItems[final.cursor]
			validSelection = profileIndex < len(profiles)
		}
	} else {
		if final.cursor < len(profiles) {
			profileIndex = final.cursor
			validSelection = true
		}
	}

	if !validSelection {
		return
	}

	selected := profiles[profileIndex]

	if final.view {
		// View file content
		if !selected.IsDir {
			viewProfileFile(selected.Path)
			// Show menu again after viewing
			chooseProfile(profilesPath, private)
		}
		return
	}

	if final.selected {
		if selected.IsDir {
			chooseProfile(selected.Path, private)
		} else {
			loadProfileFile(selected.Path)
		}
	}
}

type profileModel struct {
	choices       []string
	profiles      []Profile
	cursor        int
	selected      bool
	view          bool
	private       bool
	path          string
	searchMode    bool
	searchInput   string
	filteredItems []int
}

func (m profileModel) Init() tea.Cmd {
	return nil
}

func (m profileModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.searchMode {
			switch msg.String() {
			case "ctrl+c", "esc":
				m.searchMode = false
				m.searchInput = ""
				m.filteredItems = nil
				m.cursor = 0
			case "enter":
				if len(m.filteredItems) > 0 && m.cursor < len(m.filteredItems) {
					m.selected = true
					return m, tea.Quit
				}
			case "backspace":
				if len(m.searchInput) > 0 {
					m.searchInput = m.searchInput[:len(m.searchInput)-1]
					m.filteredItems = m.filterProfiles()
					m.cursor = 0
				}
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.filteredItems)-1 {
					m.cursor++
				}
			default:
				if len(msg.String()) == 1 {
					m.searchInput += msg.String()
					m.filteredItems = m.filterProfiles()
					m.cursor = 0
				}
			}
		} else {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "/":
				m.searchMode = true
				m.searchInput = ""
				m.filteredItems = nil
				m.cursor = 0
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
	}
	return m, nil
}

func (m profileModel) View() string {
	if m.searchMode {
		s := fmt.Sprintf("Search: %s\n\n", m.searchInput)
		if len(m.filteredItems) == 0 {
			s += "No matches found\n"
		} else {
			for i, idx := range m.filteredItems {
				cursor := " "
				if m.cursor == i {
					cursor = ">"
				}
				s += fmt.Sprintf("%s %s\n", cursor, m.choices[idx])
			}
		}
		s += "\nPress Esc to exit search, Enter to select.\n"
		return s
	}

	s := "Select Profile:\n\n"
	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}
	s += "\nPress / to search, v to view, Enter to select, q to quit.\n"
	return s
}

func (m profileModel) filterProfiles() []int {
	if m.searchInput == "" {
		// Return all profiles when search is empty
		var all []int
		for i := range m.choices {
			all = append(all, i)
		}
		return all
	}

	var matches []int
	searchLower := strings.ToLower(m.searchInput)

	for i, choice := range m.choices {
		if strings.Contains(strings.ToLower(choice), searchLower) {
			matches = append(matches, i)
		}
	}

	return matches
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
		fmt.Fprintf(os.Stderr, "Error getting namespaces: %v\n", err)
		return
	}

	namespaces := strings.Fields(string(output))
	if len(namespaces) == 0 {
		fmt.Fprintln(os.Stderr, "No namespaces found")
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

	final := finalModel.(namespaceModel)

	// Determine the actual namespace index
	var namespaceIndex int
	var validSelection bool

	if final.searchMode && len(final.filteredItems) > 0 {
		if final.cursor < len(final.filteredItems) {
			namespaceIndex = final.filteredItems[final.cursor]
			validSelection = namespaceIndex < len(namespaces)
		}
	} else {
		if final.cursor < len(namespaces) {
			namespaceIndex = final.cursor
			validSelection = true
		}
	}

	if validSelection && final.selected {
		selectedNamespace := namespaces[namespaceIndex]
		setNamespace(selectedNamespace)
	}
}

type namespaceModel struct {
	choices       []string
	cursor        int
	selected      bool
	searchMode    bool
	searchInput   string
	filteredItems []int
}

func (m namespaceModel) Init() tea.Cmd {
	return nil
}

func (m namespaceModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.searchMode {
			switch msg.String() {
			case "ctrl+c", "esc":
				m.searchMode = false
				m.searchInput = ""
				m.filteredItems = nil
				m.cursor = 0
			case "enter":
				if len(m.filteredItems) > 0 && m.cursor < len(m.filteredItems) {
					m.selected = true
					return m, tea.Quit
				}
			case "backspace":
				if len(m.searchInput) > 0 {
					m.searchInput = m.searchInput[:len(m.searchInput)-1]
					m.filteredItems = m.filterNamespaces()
					m.cursor = 0
				}
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.filteredItems)-1 {
					m.cursor++
				}
			default:
				if len(msg.String()) == 1 {
					m.searchInput += msg.String()
					m.filteredItems = m.filterNamespaces()
					m.cursor = 0
				}
			}
		} else {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "/":
				m.searchMode = true
				m.searchInput = ""
				m.filteredItems = nil
				m.cursor = 0
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
	}
	return m, nil
}

func (m namespaceModel) View() string {
	if m.searchMode {
		s := fmt.Sprintf("Search: %s\n\n", m.searchInput)
		if len(m.filteredItems) == 0 {
			s += "No matches found\n"
		} else {
			for i, idx := range m.filteredItems {
				cursor := " "
				if m.cursor == i {
					cursor = ">"
				}
				s += fmt.Sprintf("%s %s\n", cursor, m.choices[idx])
			}
		}
		s += "\nPress Esc to exit search, Enter to select.\n"
		return s
	}

	s := "Select Namespace:\n\n"
	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}
	s += "\nPress / to search, Enter to select, q to quit.\n"
	return s
}

func (m namespaceModel) filterNamespaces() []int {
	if m.searchInput == "" {
		// Return all namespaces when search is empty
		var all []int
		for i := range m.choices {
			all = append(all, i)
		}
		return all
	}

	var matches []int
	searchLower := strings.ToLower(m.searchInput)

	for i, choice := range m.choices {
		if strings.Contains(strings.ToLower(choice), searchLower) {
			matches = append(matches, i)
		}
	}

	return matches
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
