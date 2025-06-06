package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mattn/go-isatty"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

var helpText = strings.TrimSpace(`
nap is a code snippet manager for your terminal.
https://github.com/isabelroses/nap

Usage:
  nap           - for interactive mode
  nap list      - list all snippets
  nap <snippet> - print snippet to stdout

Create:
  nap < main.go                 - save snippet from stdin
  nap example/main.go < main.go - save snippet with name`)

func main() {
	runCLI(os.Args[1:])
}

func runCLI(args []string) {
	config := readConfig()
	snippets := readSnippets(config)
	snippets = migrateSnippets(config, snippets)
	snippets = scanSnippets(config, snippets)

	stdin := readStdin()
	if stdin != "" {
		saveSnippet(stdin, args, config, snippets)
		return
	}

	if len(args) > 0 {
		switch args[0] {
		case "list":
			listSnippets(snippets)
		case "-h", "--help":
			fmt.Println(helpText)
		default:
			snippet := findSnippet(args[0], snippets)
			fmt.Print(snippet.Content(isatty.IsTerminal(os.Stdout.Fd())))
		}
		return
	}

	err := runInteractiveMode(config, snippets)
	if err != nil {
		fmt.Println("Alas, there's been an error", err)
	}
}

// parseName returns a folder, name, and language for the given name.
// this is useful for parsing file names when passed as command line arguments.
//
// Example:
//
//	Notes/Hello.go -> (Notes, Hello, go)
//	Hello.go       -> (Misc, Hello, go)
//	Notes/Hello    -> (Notes, Hello, go)
func parseName(s string) (string, string, string) {
	var (
		folder    = defaultSnippetFolder
		name      = defaultSnippetName
		language  = defaultLanguage
		remaining string
	)

	tokens := strings.Split(s, "/")
	if len(tokens) > 1 {
		folder = tokens[0]
		remaining = tokens[1]
	} else {
		remaining = tokens[0]
	}

	tokens = strings.Split(remaining, ".")
	if len(tokens) > 1 {
		name = tokens[0]
		language = tokens[1]
	} else {
		name = tokens[0]
	}

	return folder, name, language
}

// readStdin returns the stdin that is piped in to the command line interface.
func readStdin() string {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return ""
	}

	if stat.Mode()&os.ModeCharDevice != 0 {
		return ""
	}

	reader := bufio.NewReader(os.Stdin)
	var b strings.Builder

	for {
		r, _, err := reader.ReadRune()
		if err != nil && err == io.EOF {
			break
		}
		_, err = b.WriteRune(r)
		if err != nil {
			return ""
		}
	}

	return b.String()
}

// readSnippets returns all the snippets read from the snippets.json file.
func readSnippets(config Config) []Snippet {
	var snippets []Snippet
	file := filepath.Join(config.Home, config.File)
	dir, err := os.ReadFile(file)
	if err != nil {
		// File does not exist, create one.
		err := os.MkdirAll(config.Home, os.ModePerm)
		if err != nil {
			fmt.Printf("Unable to create directory %s, %+v", config.Home, err)
		}
		f, err := os.Create(file)
		if err != nil {
			fmt.Printf("Unable to create file %s, %+v", file, err)
		}
		defer f.Close()
		dir = []byte("[]")
		_, _ = f.Write(dir)
	}
	err = json.Unmarshal(dir, &snippets)
	if err != nil {
		fmt.Printf("Unable to unmarshal %s file, %+v\n", file, err)
		return snippets
	}
	return snippets
}

// migrateSnippets migrates any legacy snippet <dir>-<file> format to the new <dir>/<file> format
func migrateSnippets(config Config, snippets []Snippet) []Snippet {
	var migrated bool
	for idx, snippet := range snippets {
		legacyPath := filepath.Join(config.Home, snippet.LegacyPath())
		if _, err := os.Stat(legacyPath); err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				fmt.Printf("could not access %q: %v\n", legacyPath, err)
			}
			continue
		}
		file := strings.TrimPrefix(snippet.LegacyPath(), fmt.Sprintf("%s-", snippet.Folder))
		newDir := filepath.Join(config.Home, snippet.Folder)
		newPath := filepath.Join(newDir, file)
		if err := os.MkdirAll(newDir, os.ModePerm); err != nil {
			fmt.Printf("could not create %q: %v\n", newDir, err)
			continue
		}
		if err := os.Rename(legacyPath, newPath); err != nil {
			fmt.Printf("could not move %q to %q: %v\n", legacyPath, newPath, err)
		}
		migrated = true
		snippet.File = file
		snippets[idx] = snippet
	}
	if migrated {
		writeSnippets(config, snippets)
	}
	return snippets
}

// scanSnippets scans for any new/removed snippets and adds them to snippets.json
func scanSnippets(config Config, snippets []Snippet) []Snippet {
	var modified bool
	snippetExists := func(path string) bool {
		for _, snippet := range snippets {
			if path == snippet.Path() {
				return true
			}
		}
		return false
	}

	homeEntries, err := os.ReadDir(config.Home)
	if err != nil {
		fmt.Printf("could not scan config home: %v\n", err)
		return snippets
	}

	for _, homeEntry := range homeEntries {
		if !homeEntry.IsDir() {
			continue
		}
		if strings.HasPrefix(homeEntry.Name(), ".") {
			continue
		}

		folderPath := filepath.Join(config.Home, homeEntry.Name())
		folderEntries, err := os.ReadDir(folderPath)
		if err != nil {
			fmt.Printf("could not scan %q: %v\n", folderPath, err)
			continue
		}

		for _, folderEntry := range folderEntries {
			if folderEntry.IsDir() {
				continue
			}

			snippetPath := filepath.Join(homeEntry.Name(), folderEntry.Name())
			if !snippetExists(snippetPath) {
				name := folderEntry.Name()
				ext := filepath.Ext(name)
				snippets = append(snippets, Snippet{
					Folder:   homeEntry.Name(),
					Date:     time.Now(),
					Name:     strings.TrimSuffix(name, ext),
					File:     name,
					Language: strings.TrimPrefix(ext, "."),
					Tags:     make([]string, 0),
				})
				modified = true
			}
		}
	}

	var idx int
	for _, snippet := range snippets {
		snippetPath := filepath.Join(config.Home, snippet.Path())
		if _, err := os.Stat(snippetPath); !errors.Is(err, fs.ErrNotExist) {
			snippets[idx] = snippet
			idx++
			modified = true
		}
	}
	snippets = snippets[:idx]

	if modified {
		writeSnippets(config, snippets)
	}

	return snippets
}

func saveSnippet(content string, args []string, config Config, snippets []Snippet) {
	// Save snippet to location
	name := defaultSnippetName
	if len(args) > 0 {
		name = strings.Join(args, " ")
	}

	folder, name, language := parseName(name)
	file := fmt.Sprintf("%s.%s", name, language)
	filePath := filepath.Join(config.Home, folder, file)
	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		fmt.Println("unable to create folder")
		return
	}
	err := os.WriteFile(filePath, []byte(content), 0o644)
	if err != nil {
		fmt.Println("unable to create snippet")
		return
	}

	// Add snippet metadata
	snippet := Snippet{
		Folder:   folder,
		Date:     time.Now(),
		Name:     name,
		File:     file,
		Language: language,
	}

	snippets = append([]Snippet{snippet}, snippets...)
	writeSnippets(config, snippets)
}

func writeSnippets(config Config, snippets []Snippet) {
	b, err := json.Marshal(snippets)
	if err != nil {
		fmt.Println("Could not marshal latest snippet data.", err)
		return
	}
	err = os.WriteFile(filepath.Join(config.Home, config.File), b, os.ModePerm)
	if err != nil {
		fmt.Println("Could not save snippets file.", err)
	}
}

func listSnippets(snippets []Snippet) {
	for _, snippet := range snippets {
		fmt.Println(snippet)
	}
}

func findSnippet(search string, snippets []Snippet) Snippet {
	matches := fuzzy.FindFrom(search, Snippets{snippets})
	if len(matches) > 0 {
		return snippets[matches[0].Index]
	}
	return Snippet{}
}

func runInteractiveMode(config Config, snippets []Snippet) error {
	if len(snippets) == 0 {
		// welcome to nap!
		snippets = append(snippets, defaultSnippet)
	}
	state := readState()

	folders := make(map[Folder][]list.Item)
	for _, snippet := range snippets {
		folders[Folder(snippet.Folder)] = append(folders[Folder(snippet.Folder)], list.Item(snippet))
	}

	defaultStyles := DefaultStyles(config)

	var folderItems []list.Item
	foldersSlice := maps.Keys(folders)
	slices.Sort(foldersSlice)
	for _, folder := range foldersSlice {
		folderItems = append(folderItems, list.Item(folder))
	}
	if len(folderItems) <= 0 {
		folderItems = append(folderItems, list.Item(Folder(defaultSnippetFolder)))
	}
	folderList := list.New(folderItems, folderDelegate{defaultStyles.Folders.Blurred}, 0, 0)
	folderList.Title = "Folders"

	folderList.SetShowHelp(false)
	folderList.SetFilteringEnabled(false)
	folderList.SetShowStatusBar(false)
	folderList.DisableQuitKeybindings()
	folderList.Styles.NoItems = lipgloss.NewStyle().Margin(0, 2).Foreground(lipgloss.Color(config.GrayColor))
	folderList.SetStatusBarItemName("folder", "folders")

	for idx, folder := range foldersSlice {
		if string(folder) == state.CurrentFolder {
			folderList.Select(idx)
			break
		}
	}

	content := viewport.New(80, 0)

	lists := map[Folder]*list.Model{}

	currentFolder := folderList.SelectedItem().(Folder)
	for folder, items := range folders {
		snippetList := newList(items, 20, defaultStyles.Snippets.Focused)
		if folder == currentFolder {
			for idx, item := range snippetList.Items() {
				if s, ok := item.(Snippet); ok && s.File == state.CurrentSnippet {
					snippetList.Select(idx)
					break
				}
			}
		}
		lists[folder] = snippetList
	}

	shit_help := help.New()
	shit_help.Styles = help.Styles{
		ShortKey:       lipgloss.NewStyle().Foreground(lipgloss.Color("#c6a0f6")), // Magenta
		ShortDesc:      lipgloss.NewStyle().Foreground(lipgloss.Color("#f0c6c6")), // White
		ShortSeparator: lipgloss.NewStyle().Foreground(lipgloss.Color("8")),       // Gray
		FullKey:        lipgloss.NewStyle().Foreground(lipgloss.Color("#c6a0f6")), // Magenta
		FullDesc:       lipgloss.NewStyle().Foreground(lipgloss.Color("#f0c6c6")), // White
		FullSeparator:  lipgloss.NewStyle().Foreground(lipgloss.Color("8")),       // Gray
		Ellipsis:       lipgloss.NewStyle().Foreground(lipgloss.Color("8")),       // Gray
	}

	m := &Model{
		Lists:        lists,
		Folders:      folderList,
		Code:         content,
		ContentStyle: defaultStyles.Content.Blurred,
		ListStyle:    defaultStyles.Snippets.Focused,
		FoldersStyle: defaultStyles.Folders.Blurred,
		keys:         DefaultKeyMap,
		help:         shit_help,
		config:       config,
		inputs: []textinput.Model{
			newTextInput(defaultSnippetFolder + " "),
			newTextInput(defaultSnippetName),
		},
		tagsInput: newTextInput("Tags"),
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	model, err := p.Run()
	if err != nil {
		return err
	}
	fm, ok := model.(*Model)
	if !ok {
		return err
	}
	var allSnippets []list.Item
	for _, list := range fm.Lists {
		allSnippets = append(allSnippets, list.Items()...)
	}
	b, err := json.Marshal(allSnippets)
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join(config.Home, config.File), b, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func newList(items []list.Item, height int, styles SnippetsBaseStyle) *list.Model {
	snippetList := list.New(items, snippetDelegate{styles, navigatingState}, 25, height)
	snippetList.SetShowHelp(false)
	snippetList.SetShowFilter(false)
	snippetList.SetShowTitle(false)
	snippetList.Styles.StatusBar = lipgloss.NewStyle().Margin(1, 2).Foreground(lipgloss.Color("240")).MaxWidth(35 - 2)
	snippetList.Styles.NoItems = lipgloss.NewStyle().Margin(0, 2).Foreground(lipgloss.Color("8")).MaxWidth(35 - 2)
	snippetList.FilterInput.Prompt = "Find: "
	snippetList.FilterInput.PromptStyle = styles.Title
	snippetList.SetStatusBarItemName("snippet", "snippets")
	snippetList.DisableQuitKeybindings()
	snippetList.Styles.Title = styles.Title
	snippetList.Styles.TitleBar = styles.TitleBar

	return &snippetList
}

func newTextInput(placeholder string) textinput.Model {
	i := textinput.New()
	i.Prompt = ""
	i.PromptStyle = lipgloss.NewStyle().Margin(0, 1)
	i.Placeholder = placeholder
	return i
}
