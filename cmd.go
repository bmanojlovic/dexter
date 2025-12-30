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
    eval "$(%s "$@")"
}
`, binaryPath)
		case "fish":
			fmt.Printf(`# Add this to your ~/.config/fish/config.fish:
# eval "$(dexter init fish)" >> ~/.config/fish/config.fish
function dexctx
    eval (%s $argv)
end
`, binaryPath)
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
