package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	privateFlag     bool
	groupFlag       string
	namespaceFlag   bool
	awsArg          string
	profilesRoot    string
	listAWS         bool
)

func init() {
	profilesRoot = filepath.Join(os.Getenv("HOME"), ".dexter_profiles")
}

var rootCmd = &cobra.Command{
	Use:   "dexter [profile_number]",
	Short: "AWS profile and Kubernetes namespace manager",
	Long:  "Interactive tool for managing AWS profiles and Kubernetes namespaces",
	Args:  cobra.MaximumNArgs(1),
	Version: version + " (" + commit + ") built on " + date,
	Run: func(cmd *cobra.Command, args []string) {
		// Handle -p flag with optional profile argument
		if privateFlag {
			if len(args) == 1 {
				loadProfile(profilesRoot, args[0], true)
			} else {
				chooseProfile(profilesRoot, true)
			}
			return
		}

		// Handle -g flag
		if groupFlag != "" {
			chooseProfile(filepath.Join(profilesRoot, groupFlag), true)
			return
		}

		// Handle -n flag
		if namespaceFlag {
			if len(args) == 1 {
				setNamespace(args[0])
			} else {
				selectNamespace()
			}
			return
		}

		// Handle -a flag
		if awsArg != "" {
			setAWSProfile(awsArg)
			return
		}
		if listAWS {
			selectAWSProfile()
			return
		}

		// Handle positional argument (profile number or name)
		if len(args) == 1 {
			loadProfile(profilesRoot, args[0], false)
			return
		}

		// Default: show interactive menu
		chooseProfile(profilesRoot, false)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.Flags().BoolVarP(&privateFlag, "private", "p", false, "Show private profiles (optional: provide profile name as argument)")
	rootCmd.Flags().StringVarP(&groupFlag, "group", "g", "", "Navigate to profile group")
	rootCmd.Flags().BoolVarP(&namespaceFlag, "namespace", "n", false, "Set Kubernetes namespace (optional: provide namespace as argument)")
	rootCmd.Flags().StringVarP(&awsArg, "aws", "a", "", "Set AWS profile directly")
	rootCmd.Flags().BoolVar(&listAWS, "list-aws", false, "List and select AWS profile")
}

var initCmd = &cobra.Command{
	Use:   "init [shell]",
	Short: "Output shell wrapper function",
	Long:  "Outputs the shell wrapper function to enable environment variable exports. Supported shells: bash, zsh, fish",
	Run: func(cmd *cobra.Command, args []string) {
		shell := "bash"
		if len(args) > 0 {
			shell = args[0]
		}
		
		binaryPath, _ := os.Executable()
		
		switch shell {
		case "bash", "zsh":
			fmt.Printf(`# Add this to your ~/.bashrc or ~/.zshrc:
# eval "$(dexter init)" >> ~/.bashrc
dexctx() {
    if [ "$1" = "update" ]; then
        echo "Checking for updates..." >&2
        CURRENT=$(%s --version 2>&1 | head -n1 | awk '{print $3}')
        
        LATEST=$(curl -s https://api.github.com/repos/akomic/dexter/releases/latest | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
        
        if [ -z "$LATEST" ]; then
            echo "Failed to check for updates" >&2
            return 1
        fi
        
        if [ "$LATEST" = "$CURRENT" ]; then
            echo "Already on latest version: $CURRENT" >&2
            return 0
        fi
        
        echo "Updating from $CURRENT to $LATEST..." >&2
        ARCH=$(uname -m)
        OS=$(uname -s | tr '[:upper:]' '[:lower:]')
        [ "$ARCH" = "x86_64" ] && ARCH="amd64"
        
        BINARY="dexter-${OS}-${ARCH}"
        URL="https://github.com/akomic/dexter/releases/latest/download/${BINARY}"
        
        TEMP=$(mktemp)
        if ! curl -sL "$URL" -o "$TEMP"; then
            echo "Failed to download update" >&2
            rm -f "$TEMP"
            return 1
        fi
        
        chmod +x "$TEMP"
        DEXTER_PATH=$(which dexter)
        mv "$TEMP" "$DEXTER_PATH"
        echo "Successfully updated to $LATEST" >&2
    else
        eval "$(%s "$@")"
    fi
}
`, binaryPath, binaryPath)
		case "fish":
			fmt.Printf(`# Add this to your ~/.config/fish/config.fish:
# eval "$(dexter init fish)" >> ~/.config/fish/config.fish
function dexctx
    if test "$argv[1]" = "update"
        echo "Checking for updates..." >&2
        set LATEST (curl -s https://api.github.com/repos/akomic/dexter/releases/latest | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
        set CURRENT (%s --version 2>&1 | awk '{print $1}')
        
        if test "$LATEST" != "$CURRENT"
            echo "Updating from $CURRENT to $LATEST..." >&2
            set ARCH (uname -m)
            set OS (uname -s | tr '[:upper:]' '[:lower:]')
            test "$ARCH" = "x86_64"; and set ARCH "amd64"
            
            set BINARY "dexter-$OS-$ARCH"
            set URL "https://github.com/akomic/dexter/releases/latest/download/$BINARY"
            
            set TEMP (mktemp)
            curl -sL "$URL" -o "$TEMP"
            chmod +x "$TEMP"
            
            set DEXTER_PATH (which dexter)
            mv "$TEMP" "$DEXTER_PATH"
            echo "Updated to $LATEST" >&2
        else
            echo "Already on latest version: $CURRENT" >&2
        end
    else
        eval (%s $argv)
    end
end
`, binaryPath, binaryPath)
		default:
			fmt.Fprintf(os.Stderr, "Unsupported shell: %s\n", shell)
		}
	},
}

func main() {
	rootCmd.SetOut(os.Stderr)
	rootCmd.SetErr(os.Stderr)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
