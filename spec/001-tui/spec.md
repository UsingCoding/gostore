# Gostore TUI feature spec

This documents describe spec for TUI for gostore

## Spec

TUI allows users to:
* View secrets in tree list
* Examine each secret payload
* Manage secret fields (add/remove)
* Edit secret via $EDITOR (using existed feature `internal/gostore/app/usecase/edit/editor.go`)
* Copy selected secret field to buffer
* Switch between stores (using existed `internal/gostore/app/config/service.go`)
* Manage stores (remove)

### Requirements to TUI

* Responsive as much as possible
* Focused widget should be highlighted by using different color (green border for example)
* Selected rows in non-focused lists should use a light blue highlight to differentiate from focused selections
* Use `github.com/metaspartan/gotui/v5`
* Each box corners should be rounded (using gotui features)

### Components

Here described TUI elements in yaml-style

```yaml
screen: dashboard
# Global hotkeys
hotkeys:
  - keys: 
      - <q>
      - <C-c>
    action: exit
panels:
  - id: sidebar
    description: |
      Located at left side
      Each element of sidebar numerated (1..N), by pressing specific number UI focuses specific widget in sidebar
      Takes 1/4 of space by horizontal alongside other content
      In vertical - takes all place
      This sidebar used only for separation and may not be presented in code 
    children:
      - id: sidebar.info
        description: |
          Box, named `Context`
          In box only name of current store
          Vertically takes only one line to fit store name. Store names only one-lined
      - id: sidebar.secrets-list
        note: use tree widget
        description: |
          Collapsable list of store entries from `internal/gostore/app/store/service.go:List`
          Elements collapsed by default
          Vertically takes all available place 
        hotkeys:
          - key: <Space>
            action: Collapse/ncollapse tree entry
          - key: </>
            action: Activate search bar and focuses on it
          - key: <e>
            action: edit entry via `internal/gostore/app/usecase/edit/editor.go`
          - key: <d>
            action: delete entry
        children:
          - id: sidebar.secrets-list.search-bar
            description: |
              One line search bar
      - id: sidebar.stores-list
        description: |
          Simple list of stores from `internal/gostore/app/config/service.go:ListStores`
          Vertically takes about 1/3 of available space and keeps at least 3 visible rows
          Active store in list starts with `*`
        hotkeys:
          - key: <Space>
            action: Switch active store
          - key: </>
            action: Activate search bar and focuses on it
        children:
          - id: sidebar.secrets-list.search-bar
            description: |
              One line search bar
  - id: secret-pane
    description: |
      Contains boxes for each secret field
    children:
      - id: secret-pane.secret-field
        notes: 
          - Wrap with scrollbar, since secret may have many fields
          - Check for printable symbols when print user secret payload
        description: |
          By default field payload is hidden with `***`.
          If payload is hidden, size of payload box is limited for 1 line
          If payload is viewable (not hidden by hotkey), size of payload box is unlimited
          This child vertically repeated for each secret field
        hotkeys:
          - key: <Space>
            note: Use `github.com/atotto/clipboard` to copy to clipboard
            action: copy payload to clipboard
          - key: <e>
            action: Edit secret with `internal/gostore/app/usecase/edit/editor.go`
          - key: <v>
            action: Make payload viewable
  - id: status-bar
    description: |
      Bar with hotkeys help for current focused pane
      If error occurred - shows there too
```