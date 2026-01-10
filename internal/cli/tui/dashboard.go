package tui

import (
	"context"
	"errors"
	"fmt"
	"image"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"

	ui "github.com/metaspartan/gotui/v5"
	"github.com/metaspartan/gotui/v5/widgets"

	"github.com/UsingCoding/gostore/internal/common/maybe"
	"github.com/UsingCoding/gostore/internal/gostore/app/config"
	"github.com/UsingCoding/gostore/internal/gostore/app/storage"
	"github.com/UsingCoding/gostore/internal/gostore/app/store"
	"github.com/UsingCoding/gostore/internal/gostore/app/usecase/edit"
)

type focusArea int

const (
	focusContext focusArea = iota
	focusSecretsList
	focusSecretsSearch
	focusStoresList
	focusStoresSearch
	focusSecretPane
)

type dashboard struct {
	ui.Block

	ctx context.Context

	configService config.Service
	storeService  store.Service
	editService   edit.Service
	clipboard     *mockClipboard

	grid         *ui.Grid
	sidebar      *widgets.Flex
	secretsPanel *widgets.Flex
	storesPanel  *widgets.Flex

	infoBox       *widgets.Paragraph
	secretsTree   *widgets.Tree
	secretsSearch *widgets.Input
	storesList    *widgets.List
	storesSearch  *widgets.Input
	secretPane    *SecretPane

	focus            focusArea
	lastSidebarFocus focusArea

	normalBorder ui.Style
	focusBorder  ui.Style

	secretsTreeData storage.Tree
	secretsNodes    []*widgets.TreeNode
	secretsFilter   string

	stores         []config.StoreView
	filteredStores []config.StoreView
	storesFilter   string

	selectedSecretPath string

	modal *confirmModal
	input *textPrompt

	status string
}

type confirmModal struct {
	modal     *widgets.Modal
	onConfirm func()
	onCancel  func()
}

type textPrompt struct {
	block  ui.Block
	title  string
	prompt string
	input  *widgets.Input

	onSubmit func(string)
	onCancel func()
}

type mockClipboard struct {
	value string
}

func (c *mockClipboard) Copy(payload string) {
	c.value = payload
}

func newDashboard(ctx context.Context, configService config.Service, storeService store.Service, editService edit.Service) *dashboard {
	d := &dashboard{
		Block:            *ui.NewBlock(),
		ctx:              ctx,
		configService:    configService,
		storeService:     storeService,
		editService:      editService,
		clipboard:        &mockClipboard{},
		normalBorder:     ui.Theme.Block.Border,
		focusBorder:      ui.NewStyle(ui.ColorGreen),
		focus:            focusSecretsList,
		lastSidebarFocus: focusSecretsList,
	}
	d.Border = false
	d.initWidgets()
	d.refreshAll()

	if d.editService == nil {
		d.setStatus("Editor unavailable")
	}

	return d
}

func (d *dashboard) initWidgets() {
	d.infoBox = widgets.NewParagraph()
	d.infoBox.Title = "1 - Context"
	d.infoBox.WrapText = false
	d.infoBox.BorderRounded = true

	d.secretsTree = widgets.NewTree()
	d.secretsTree.Title = "2 - Secrets"
	d.secretsTree.WrapText = false
	d.secretsTree.SelectedRowStyle = ui.NewStyle(ui.ColorBlack, ui.ColorGreen)
	d.secretsTree.BorderRounded = true

	d.secretsSearch = widgets.NewInput()
	d.secretsSearch.Title = "Search"
	d.secretsSearch.Placeholder = "Press / to search"
	d.secretsSearch.BorderRounded = true

	d.storesList = widgets.NewList()
	d.storesList.Title = "3 - Stores"
	d.storesList.WrapText = false
	d.storesList.SelectedStyle = ui.NewStyle(ui.ColorBlack, ui.ColorGreen)
	d.storesList.BorderRounded = true

	d.storesSearch = widgets.NewInput()
	d.storesSearch.Title = "Search"
	d.storesSearch.Placeholder = "Press / to search"
	d.storesSearch.BorderRounded = true

	d.secretPane = NewSecretPane()

	d.secretsPanel = widgets.NewFlex()
	d.secretsPanel.Border = false
	d.secretsPanel.Direction = widgets.FlexColumn
	d.secretsPanel.AddItem(d.secretsTree, 0, 1, false)
	d.secretsPanel.AddItem(d.secretsSearch, 3, 0, false)

	d.storesPanel = widgets.NewFlex()
	d.storesPanel.Border = false
	d.storesPanel.Direction = widgets.FlexColumn
	d.storesPanel.AddItem(d.storesList, 0, 1, false)
	d.storesPanel.AddItem(d.storesSearch, 3, 0, false)

	d.sidebar = widgets.NewFlex()
	d.sidebar.Border = false
	d.sidebar.Direction = widgets.FlexColumn
	d.sidebar.AddItem(d.infoBox, 3, 0, false)
	d.sidebar.AddItem(d.secretsPanel, 0, 5, false)
	d.sidebar.AddItem(d.storesPanel, 0, 1, false)

	d.grid = ui.NewGrid()
	d.grid.Set(
		ui.NewCol(0.25, d.sidebar),
		ui.NewCol(0.75, d.secretPane),
	)

	d.applyFocusStyles()
}

func (d *dashboard) Draw(buf *ui.Buffer) {
	d.applyFocusStyles()
	d.secretPane.SetStatus(d.status)

	d.grid.SetRect(d.Min.X, d.Min.Y, d.Max.X, d.Max.Y)
	d.grid.Draw(buf)

	if d.modal != nil {
		d.modal.modal.BorderRounded = true
		d.modal.modal.CenterIn(d.Min.X, d.Min.Y, d.Max.X, d.Max.Y, 60, 9)
		d.modal.modal.Draw(buf)
	}

	if d.input != nil {
		d.input.CenterIn(d.Min.X, d.Min.Y, d.Max.X, d.Max.Y, 60, 9)
		d.input.Draw(buf)
	}
}

func (d *dashboard) HandleEvent(e ui.Event) bool {
	if e.Type != ui.KeyboardEvent {
		return false
	}

	if d.modal != nil {
		return d.handleModalEvent(e)
	}

	if d.input != nil {
		return d.handleInputEvent(e)
	}

	if d.handleGlobalKeys(e) {
		return true
	}

	switch d.focus {
	case focusSecretsList:
		return d.handleSecretsListEvent(e)
	case focusSecretsSearch:
		return d.handleSearchEvent(e, focusSecretsSearch)
	case focusStoresList:
		return d.handleStoresListEvent(e)
	case focusStoresSearch:
		return d.handleSearchEvent(e, focusStoresSearch)
	case focusSecretPane:
		return d.handleSecretPaneEvent(e)
	case focusContext:
		return false
	default:
		return false
	}
}

func (d *dashboard) handleGlobalKeys(e ui.Event) bool {
	switch e.ID {
	case "1":
		d.setFocus(focusContext)
		return true
	case "2":
		d.setFocus(focusSecretsList)
		return true
	case "3":
		d.setFocus(focusStoresList)
		return true
	case "<Tab>":
		if d.focus == focusSecretPane {
			d.setFocus(d.lastSidebarFocus)
		} else {
			d.setFocus(focusSecretPane)
		}
		return true
	}
	return false
}

func (d *dashboard) handleSecretsListEvent(e ui.Event) bool {
	if e.ID == "/" {
		d.startSearch(focusSecretsSearch)
		return true
	}

	if d.secretsTree.SelectedNode() == nil {
		return false
	}

	switch e.ID {
	case "<Space>", " ":
		d.secretsTree.ToggleExpand()
		d.updateSelectedSecret()
		return true
	case "e":
		d.editSelectedSecret()
		return true
	case "d":
		d.confirmRemoveSecret()
		return true
	case "j", "<Down>":
		d.secretsTree.ScrollDown()
		d.updateSelectedSecret()
		return true
	case "k", "<Up>":
		d.secretsTree.ScrollUp()
		d.updateSelectedSecret()
		return true
	case "<Home>":
		d.secretsTree.ScrollTop()
		d.updateSelectedSecret()
		return true
	case "<End>":
		d.secretsTree.ScrollBottom()
		d.updateSelectedSecret()
		return true
	case "<Enter>":
		d.secretsTree.ToggleExpand()
		return true
	}

	return false
}

func (d *dashboard) handleStoresListEvent(e ui.Event) bool {
	if e.ID == "/" {
		d.startSearch(focusStoresSearch)
		return true
	}

	if len(d.filteredStores) == 0 {
		return false
	}

	switch e.ID {
	case "<Space>", " ":
		d.switchStore()
		return true
	case "d":
		d.confirmRemoveStore()
		return true
	case "j", "<Down>":
		d.storesList.ScrollDown()
		return true
	case "k", "<Up>":
		d.storesList.ScrollUp()
		return true
	case "<Home>":
		d.storesList.ScrollTop()
		return true
	case "<End>":
		d.storesList.ScrollBottom()
		return true
	}

	return false
}

func (d *dashboard) handleSecretPaneEvent(e ui.Event) bool {
	switch e.ID {
	case "j", "<Down>":
		d.secretPane.MoveSelection(1)
		return true
	case "k", "<Up>":
		d.secretPane.MoveSelection(-1)
		return true
	case "<Space>", " ":
		d.copySelectedField()
		return true
	case "v":
		d.secretPane.ToggleVisible()
		return true
	case "e":
		d.editSelectedField()
		return true
	case "d":
		d.confirmRemoveField()
		return true
	case "a":
		d.promptAddField()
		return true
	}

	return false
}

func (d *dashboard) handleSearchEvent(e ui.Event, target focusArea) bool {
	switch e.ID {
	case "<Enter>":
		d.applySearch(target)
		return true
	case "<Esc>":
		d.cancelSearch(target)
		return true
	}

	switch target {
	case focusSecretsSearch:
		return handleInputKey(d.secretsSearch, e)
	case focusStoresSearch:
		return handleInputKey(d.storesSearch, e)
	default:
		return false
	}
}

func (d *dashboard) handleModalEvent(e ui.Event) bool {
	modal := d.modal
	if modal == nil {
		return false
	}

	switch e.ID {
	case "<Left>", "h":
		if modal.modal.ActiveButtonIndex > 0 {
			modal.modal.ActiveButtonIndex--
		}
		return true
	case "<Right>", "l":
		if modal.modal.ActiveButtonIndex < len(modal.modal.Buttons)-1 {
			modal.modal.ActiveButtonIndex++
		}
		return true
	case "<Tab>":
		if len(modal.modal.Buttons) > 0 {
			modal.modal.ActiveButtonIndex = (modal.modal.ActiveButtonIndex + 1) % len(modal.modal.Buttons)
		}
		return true
	case "<Enter>":
		if modal.modal.ActiveButtonIndex == 0 && modal.onConfirm != nil {
			modal.onConfirm()
		} else if modal.onCancel != nil {
			modal.onCancel()
		}
		d.modal = nil
		return true
	case "<Esc>":
		if modal.onCancel != nil {
			modal.onCancel()
		}
		d.modal = nil
		return true
	}

	return true
}

func (d *dashboard) handleInputEvent(e ui.Event) bool {
	if d.input == nil {
		return false
	}

	switch e.ID {
	case "<Enter>":
		value := strings.TrimSpace(d.input.input.Text)
		if value != "" {
			d.input.onSubmit(value)
		} else {
			d.setStatus("Value required")
		}
		d.input = nil
		d.setFocus(focusSecretPane)
		return true
	case "<Esc>":
		if d.input.onCancel != nil {
			d.input.onCancel()
		}
		d.input = nil
		d.setFocus(focusSecretPane)
		return true
	}

	return handleInputKey(d.input.input, e)
}

func handleInputKey(input *widgets.Input, e ui.Event) bool {
	switch e.ID {
	case "<Backspace>":
		input.Backspace()
		return true
	case "<Left>":
		input.MoveCursorLeft()
		return true
	case "<Right>":
		input.MoveCursorRight()
		return true
	case "<Space>", " ":
		input.InsertRune(' ')
		return true
	}

	if r, ok := inputRune(e.ID); ok {
		input.InsertRune(r)
		return true
	}

	return false
}

func inputRune(id string) (rune, bool) {
	if strings.HasPrefix(id, "<") && strings.HasSuffix(id, ">") {
		return 0, false
	}
	runes := []rune(id)
	if len(runes) != 1 {
		return 0, false
	}
	return runes[0], true
}

func (d *dashboard) startSearch(target focusArea) {
	switch target {
	case focusSecretsSearch:
		setInputText(d.secretsSearch, d.secretsFilter)
		d.setFocus(focusSecretsSearch)
	case focusStoresSearch:
		setInputText(d.storesSearch, d.storesFilter)
		d.setFocus(focusStoresSearch)
	}
}

func (d *dashboard) applySearch(target focusArea) {
	switch target {
	case focusSecretsSearch:
		d.secretsFilter = strings.ToLower(strings.TrimSpace(d.secretsSearch.Text))
		d.applySecretsFilter()
		d.setFocus(focusSecretsList)
		d.updateSelectedSecret()
	case focusStoresSearch:
		d.storesFilter = strings.ToLower(strings.TrimSpace(d.storesSearch.Text))
		d.applyStoresFilter()
		d.setFocus(focusStoresList)
	}
}

func (d *dashboard) cancelSearch(target focusArea) {
	switch target {
	case focusSecretsSearch:
		setInputText(d.secretsSearch, d.secretsFilter)
		d.setFocus(focusSecretsList)
	case focusStoresSearch:
		setInputText(d.storesSearch, d.storesFilter)
		d.setFocus(focusStoresList)
	}
}

func setInputText(input *widgets.Input, text string) {
	input.Text = text
	input.Cursor = len([]rune(text))
}

func (d *dashboard) setFocus(focus focusArea) {
	d.focus = focus
	switch focus {
	case focusContext, focusSecretsList, focusStoresList:
		d.lastSidebarFocus = focus
	}
	d.applyFocusStyles()
}

func (d *dashboard) applyFocusStyles() {
	d.setBorder(&d.infoBox.Block, d.focus == focusContext)
	d.setBorder(&d.secretsTree.Block, d.focus == focusSecretsList)
	d.setBorder(&d.secretsSearch.Block, d.focus == focusSecretsSearch)
	d.setBorder(&d.storesList.Block, d.focus == focusStoresList)
	d.setBorder(&d.storesSearch.Block, d.focus == focusStoresSearch)
	d.secretPane.SetFocused(d.focus == focusSecretPane)
}

func (d *dashboard) setBorder(block *ui.Block, focused bool) {
	if focused {
		block.BorderStyle = d.focusBorder
	} else {
		block.BorderStyle = d.normalBorder
	}
	block.BorderRounded = true
}

func (d *dashboard) refreshAll() {
	d.refreshContext()
	d.refreshStores()
	d.refreshSecrets(false)
}

func (d *dashboard) refreshContext() {
	storeID, err := d.configService.CurrentStoreID(d.ctx)
	if err != nil {
		d.infoBox.Text = "No store"
		d.setStatus(fmt.Sprintf("Failed to load context: %v", err))
		return
	}

	if id, ok := maybe.JustValid(storeID); ok {
		d.infoBox.Text = string(id)
		return
	}

	d.infoBox.Text = "No store"
}

func (d *dashboard) refreshStores() {
	stores, err := d.configService.ListStores(d.ctx)
	if err != nil {
		d.setStatus(fmt.Sprintf("Failed to list stores: %v", err))
		d.storesList.Rows = nil
		d.filteredStores = nil
		return
	}

	d.stores = stores
	d.applyStoresFilter()
}

func (d *dashboard) applyStoresFilter() {
	query := strings.ToLower(strings.TrimSpace(d.storesFilter))
	if query == "" {
		d.filteredStores = d.stores
	} else {
		d.filteredStores = nil
		for _, store := range d.stores {
			if strings.Contains(strings.ToLower(string(store.ID)), query) {
				d.filteredStores = append(d.filteredStores, store)
			}
		}
	}

	d.storesList.Rows = storeRows(d.filteredStores)
	if len(d.storesList.Rows) == 0 {
		d.storesList.SelectedRow = 0
		return
	}
	if d.storesList.SelectedRow >= len(d.storesList.Rows) {
		d.storesList.SelectedRow = len(d.storesList.Rows) - 1
	}
}

func storeRows(stores []config.StoreView) []string {
	rows := make([]string, 0, len(stores))
	for _, store := range stores {
		prefix := "  "
		if store.Current {
			prefix = "* "
		}
		rows = append(rows, prefix+string(store.ID))
	}
	return rows
}

func (d *dashboard) refreshSecrets(preserveSelection bool) {
	prev := d.selectedSecretPath

	tree, err := d.storeService.List(d.ctx, store.ListParams{})
	if err != nil {
		d.setStatus(fmt.Sprintf("Failed to list secrets: %v", err))
		d.secretsTreeData = nil
		d.secretsNodes = nil
		d.secretsTree.SetNodes(nil)
		d.selectedSecretPath = ""
		d.secretPane.SetFields(nil, "Select a secret")
		d.secretPane.TitleBottomLeft = ""
		return
	}

	d.secretsTreeData = tree
	d.applySecretsFilter()

	if preserveSelection && prev != "" {
		d.setTreeSelectionByPath(prev)
	}

	d.updateSelectedSecret()
}

func (d *dashboard) applySecretsFilter() {
	query := strings.ToLower(strings.TrimSpace(d.secretsFilter))
	filtered := filterTree(d.secretsTreeData, "", query)
	d.secretsNodes = buildTreeNodes(filtered, "")
	d.secretsTree.SetNodes(d.secretsNodes)
	d.secretsTree.SelectedRow = 0
}

func (d *dashboard) updateSelectedSecret() {
	path, ok := d.selectedSecretFromTree()
	if !ok {
		d.selectedSecretPath = ""
		d.secretPane.SetFields(nil, "Select a secret")
		d.secretPane.TitleBottomLeft = ""
		return
	}

	if path == d.selectedSecretPath {
		return
	}

	d.selectedSecretPath = path
	d.loadSecretFields(path)
}

func (d *dashboard) selectedSecretFromTree() (string, bool) {
	node := d.secretsTree.SelectedNode()
	if node == nil {
		return "", false
	}
	value, ok := node.Value.(*treeValue)
	if !ok || value == nil || !value.leaf {
		return "", false
	}
	return value.path, true
}

func (d *dashboard) loadSecretFields(path string) {
	d.secretPane.TitleBottomLeft = path

	data, err := d.storeService.Get(d.ctx, store.GetParams{
		SecretIndex: store.SecretIndex{Path: path},
	})
	if err != nil {
		d.setStatus(fmt.Sprintf("Failed to load secret: %v", err))
		d.secretPane.SetFields(nil, "Failed to load secret")
		return
	}

	fields := buildSecretFields(data)
	placeholder := ""
	if len(fields) == 0 {
		placeholder = "No printable fields"
	}

	d.secretPane.SetFields(fields, placeholder)
}

func buildSecretFields(data []store.SecretData) []secretField {
	fields := make([]secretField, 0, len(data))
	for _, entry := range data {
		if !isPrintablePayload(entry.Payload) {
			continue
		}
		fields = append(fields, secretField{
			name:         entry.Name,
			payload:      string(entry.Payload),
			defaultField: entry.Default,
		})
	}
	return fields
}

func isPrintablePayload(payload []byte) bool {
	if len(payload) == 0 {
		return true
	}
	if !utf8.Valid(payload) {
		return false
	}
	for _, r := range string(payload) {
		switch r {
		case '\n', '\r', '\t':
			continue
		default:
			if !unicode.IsPrint(r) {
				return false
			}
		}
	}
	return true
}

func (d *dashboard) setTreeSelectionByPath(path string) {
	if path == "" || len(d.secretsNodes) == 0 {
		return
	}
	flat := flattenTreeNodes(d.secretsNodes, nil)
	for i, node := range flat {
		value, ok := node.Value.(*treeValue)
		if ok && value != nil && value.path == path {
			d.secretsTree.SelectedRow = i
			return
		}
	}
}

func flattenTreeNodes(nodes []*widgets.TreeNode, out []*widgets.TreeNode) []*widgets.TreeNode {
	for _, node := range nodes {
		out = append(out, node)
		if node.Expanded {
			out = flattenTreeNodes(node.Nodes, out)
		}
	}
	return out
}

func (d *dashboard) editSelectedSecret() {
	if d.editService == nil {
		d.setStatus("Editor unavailable")
		return
	}
	path, ok := d.selectedSecretFromTree()
	if !ok {
		d.setStatus("Select a secret")
		return
	}

	err := d.editService.Edit(d.ctx, store.SecretIndex{Path: path})
	if err != nil {
		if errors.Is(err, edit.ErrNoChangesMade) {
			d.setStatus("No changes made")
			return
		}
		d.setStatus(fmt.Sprintf("Edit failed: %v", err))
		return
	}

	d.setStatus("Secret updated")
	d.refreshSecrets(true)
}

func (d *dashboard) editSelectedField() {
	if d.editService == nil {
		d.setStatus("Editor unavailable")
		return
	}
	path, ok := d.ensureSecretSelected()
	if !ok {
		return
	}
	field, ok := d.secretPane.SelectedField()
	if !ok {
		d.setStatus("Select a field")
		return
	}

	err := d.editService.Edit(d.ctx, store.SecretIndex{
		Path: path,
		Key:  maybe.NewJust(field.name),
	})
	if err != nil {
		if errors.Is(err, edit.ErrNoChangesMade) {
			d.setStatus("No changes made")
			return
		}
		d.setStatus(fmt.Sprintf("Edit failed: %v", err))
		return
	}

	d.setStatus("Field updated")
	d.loadSecretFields(path)
}

func (d *dashboard) confirmRemoveSecret() {
	path, ok := d.selectedSecretFromTree()
	if !ok {
		d.setStatus("Select a secret")
		return
	}

	d.openConfirm(fmt.Sprintf("Delete secret %s?", path), func() {
		d.removeSecret(path)
	})
}

func (d *dashboard) removeSecret(path string) {
	err := d.storeService.Remove(d.ctx, store.RemoveParams{Path: path})
	if err != nil {
		d.setStatus(fmt.Sprintf("Delete failed: %v", err))
		return
	}

	d.setStatus("Secret deleted")
	d.refreshSecrets(false)
}

func (d *dashboard) confirmRemoveField() {
	path, ok := d.ensureSecretSelected()
	if !ok {
		return
	}
	field, ok := d.secretPane.SelectedField()
	if !ok {
		d.setStatus("Select a field")
		return
	}

	d.openConfirm(fmt.Sprintf("Delete field %s?", field.name), func() {
		d.removeField(path, field.name)
	})
}

func (d *dashboard) removeField(path, key string) {
	err := d.storeService.Remove(d.ctx, store.RemoveParams{
		Path: path,
		Key:  maybe.NewJust(key),
	})
	if err != nil {
		d.setStatus(fmt.Sprintf("Remove failed: %v", err))
		return
	}

	d.setStatus("Field removed")
	d.refreshSecrets(true)
}

func (d *dashboard) promptAddField() {
	if d.editService == nil {
		d.setStatus("Editor unavailable")
		return
	}
	path, ok := d.ensureSecretSelected()
	if !ok {
		return
	}

	prompt := newTextPrompt(
		"New Field",
		"Field name:",
		"name",
		func(value string) {
			d.addField(path, value)
		},
	)
	setInputText(prompt.input, "")
	d.input = prompt
}

func (d *dashboard) addField(path, key string) {
	for _, field := range d.secretPane.fields {
		if field.name == key {
			d.setStatus("Field already exists")
			return
		}
	}

	err := d.editService.Edit(d.ctx, store.SecretIndex{
		Path: path,
		Key:  maybe.NewJust(key),
	})
	if err != nil {
		if errors.Is(err, edit.ErrNoChangesMade) {
			d.setStatus("No changes made")
			return
		}
		d.setStatus(fmt.Sprintf("Add field failed: %v", err))
		return
	}

	d.setStatus("Field added")
	d.loadSecretFields(path)
}

func (d *dashboard) copySelectedField() {
	field, ok := d.secretPane.SelectedField()
	if !ok {
		return
	}

	d.clipboard.Copy(field.payload)
	d.setStatus(fmt.Sprintf("Copied %s", field.name))
}

func (d *dashboard) switchStore() {
	storeView, ok := d.selectedStore()
	if !ok {
		d.setStatus("No store selected")
		return
	}

	err := d.configService.SetCurrentStore(d.ctx, string(storeView.ID))
	if err != nil {
		d.setStatus(fmt.Sprintf("Switch failed: %v", err))
		return
	}

	d.setStatus("Store switched")
	d.refreshAll()
}

func (d *dashboard) confirmRemoveStore() {
	storeView, ok := d.selectedStore()
	if !ok {
		d.setStatus("No store selected")
		return
	}

	d.openConfirm(fmt.Sprintf("Remove store %s?", storeView.ID), func() {
		d.removeStore(storeView.ID)
	})
}

func (d *dashboard) removeStore(id config.StoreID) {
	err := d.configService.RemoveStore(d.ctx, id)
	if err != nil {
		d.setStatus(fmt.Sprintf("Remove failed: %v", err))
		return
	}

	d.setStatus("Store removed")
	d.refreshAll()
}

func (d *dashboard) selectedStore() (config.StoreView, bool) {
	if len(d.filteredStores) == 0 {
		return config.StoreView{}, false
	}
	idx := d.storesList.SelectedRow
	if idx < 0 || idx >= len(d.filteredStores) {
		return config.StoreView{}, false
	}
	return d.filteredStores[idx], true
}

func (d *dashboard) ensureSecretSelected() (string, bool) {
	if d.selectedSecretPath == "" {
		d.setStatus("Select a secret")
		return "", false
	}
	return d.selectedSecretPath, true
}

func (d *dashboard) openConfirm(text string, onConfirm func()) {
	modal := widgets.NewModal(text)
	modal.Title = "Confirm"
	modal.BorderRounded = true
	modal.AddButton("Yes", nil)
	modal.AddButton("No", nil)
	d.modal = &confirmModal{
		modal:     modal,
		onConfirm: onConfirm,
	}
}

func (p *textPrompt) CenterIn(x1, y1, x2, y2, width, height int) {
	totalW := x2 - x1
	totalH := y2 - y1
	if width > totalW {
		width = totalW
	}
	if height > totalH {
		height = totalH
	}
	px := x1 + (totalW-width)/2
	py := y1 + (totalH-height)/2
	p.block.SetRect(px, py, px+width, py+height)
}

func (p *textPrompt) Draw(buf *ui.Buffer) {
	p.block.Draw(buf)
	if p.block.Inner.Dx() <= 0 || p.block.Inner.Dy() <= 0 {
		return
	}
	promptY := p.block.Inner.Min.Y
	buf.SetString(p.prompt, ui.NewStyle(ui.ColorWhite), image.Pt(p.block.Inner.Min.X, promptY))
	inputY := promptY + 2
	inputHeight := 3
	if inputY+inputHeight > p.block.Inner.Max.Y {
		inputHeight = p.block.Inner.Max.Y - inputY
	}
	if inputHeight <= 0 {
		return
	}
	p.input.SetRect(p.block.Inner.Min.X, inputY, p.block.Inner.Max.X, inputY+inputHeight)
	p.input.Draw(buf)
}

func newTextPrompt(title, prompt, placeholder string, onSubmit func(string)) *textPrompt {
	p := &textPrompt{
		block:    *ui.NewBlock(),
		title:    title,
		prompt:   prompt,
		input:    widgets.NewInput(),
		onSubmit: onSubmit,
	}
	p.block.Title = title
	p.block.BorderRounded = true
	p.input.BorderRounded = true
	p.input.Placeholder = placeholder
	return p
}

func (d *dashboard) setStatus(msg string) {
	d.status = msg
	d.secretPane.SetStatus(msg)
}

type treeValue struct {
	name string
	path string
	leaf bool
}

func (t *treeValue) String() string {
	return t.name
}

func buildTreeNodes(entries []storage.Entry, base string) []*widgets.TreeNode {
	nodes := make([]*widgets.TreeNode, 0, len(entries))
	for _, entry := range entries {
		path := filepath.Join(base, entry.Name)
		node := &widgets.TreeNode{
			Value:    &treeValue{name: entry.Name, path: path, leaf: len(entry.Children) == 0},
			Expanded: false,
		}
		if len(entry.Children) > 0 {
			node.Nodes = buildTreeNodes(entry.Children, path)
		}
		nodes = append(nodes, node)
	}
	return nodes
}

func filterTree(entries []storage.Entry, base string, query string) []storage.Entry {
	if query == "" {
		return entries
	}
	filtered := make([]storage.Entry, 0, len(entries))
	for _, entry := range entries {
		path := filepath.Join(base, entry.Name)
		children := filterTree(entry.Children, path, query)
		match := strings.Contains(strings.ToLower(path), query)
		if match || len(children) > 0 {
			filtered = append(filtered, storage.Entry{
				Name:     entry.Name,
				Children: children,
			})
		}
	}
	return filtered
}
