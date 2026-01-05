# 🔪 Dexter

Your friendly neighborhood profile manager. Interactive CLI tool for AWS profiles and Kubernetes namespaces with TUI menus and file validation.

## Features

- 📁 Hierarchical profile organization
- 🔒 Private profile support
- ✅ Automatic file validation (AWS config, kubeconfig)
- 👀 View profile contents before loading
- ⌨️  Interactive TUI with arrow key navigation
- 🔍 Fuzzy search for profiles and namespaces
- 🐚 Seamless shell integration

## Installation

### Download Binary

Download the latest release for your platform:

```bash
# macOS (Apple Silicon)
curl -L https://github.com/akomic/dexter/releases/latest/download/dexter-darwin-arm64 -o dexter
chmod +x dexter
sudo mv dexter /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/akomic/dexter/releases/latest/download/dexter-darwin-amd64 -o dexter
chmod +x dexter
sudo mv dexter /usr/local/bin/

# Linux (amd64)
curl -L https://github.com/akomic/dexter/releases/latest/download/dexter-linux-amd64 -o dexter
chmod +x dexter
sudo mv dexter /usr/local/bin/

# Linux (arm64)
curl -L https://github.com/akomic/dexter/releases/latest/download/dexter-linux-arm64 -o dexter
chmod +x dexter
sudo mv dexter /usr/local/bin/
```

### Build from Source

```bash
git clone https://github.com/akomic/dexter.git
cd dexter
go build
sudo mv dexter /usr/local/bin/
```

## Shell Integration

To enable environment variable exports, add the wrapper function to your shell:

```bash
# For bash/zsh
eval "$(dexter init)" >> ~/.bashrc  # or ~/.zshrc

# For fish
eval "$(dexter init fish)" >> ~/.config/fish/config.fish

# Then reload your shell
source ~/.bashrc  # or source ~/.zshrc
```

## Profile Setup

Create your profile directory structure:

```bash
mkdir -p ~/.dexter_profiles
```

Create profile files with environment variables:

```bash
# ~/.dexter_profiles/production
export AWS_PROFILE=prod
export AWS_REGION=us-east-1
export AWS_CONFIG_FILE=~/.aws/config-prod
export KUBECONFIG=~/.kube/config-prod

unset AWS_DEFAULT_REGION
```

For private profiles, add a `.private` marker:

```bash
mkdir -p ~/.dexter_profiles/company
touch ~/.dexter_profiles/company/.private
```

## Usage

```bash
# Interactive menu
dexctx

# Show private profiles
dexctx -p

# Load specific profile
dexctx -p production

# Navigate to group
dexctx -g company

# Set Kubernetes namespace
dexctx -n production

# Interactive namespace selection
dexctx -n
```

### Keyboard Shortcuts

- `↑/↓` or `j/k` - Navigate
- `Enter` - Select profile
- `v` - View profile contents
- `/` - Start fuzzy search (profiles and namespaces)
- `Esc` - Exit search mode (when searching)
- `q` - Quit

## Profile Organization

```
~/.dexter_profiles/
├── production              # Regular profile
├── staging                 # Regular profile
├── company/                # Profile group
│   ├── .private           # Marks as private
│   ├── prod
│   └── dev
└── client/
    ├── profile1
    └── profile2
```

## License

MIT
