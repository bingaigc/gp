package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	Version   = "2.0.0"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示版本信息",
	Long:  "显示CAMP-I Alpha Detector的版本信息、构建时间和系统信息",
	Run:   runVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func runVersion(cmd *cobra.Command, args []string) {
	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║   CAMP-I Distributed Alpha Signal Detector               ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("Version:      %s\n", Version)
	fmt.Printf("Build Time:   %s\n", BuildTime)
	fmt.Printf("Git Commit:   %s\n", GitCommit)
	fmt.Println()
	fmt.Printf("Go Version:   %s\n", runtime.Version())
	fmt.Printf("OS/Arch:      %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("CPUs:         %d\n", runtime.NumCPU())
	fmt.Println()
	fmt.Println("Features:")
	fmt.Println("  ✅ Real-time market data (EastMoney API)")
	fmt.Println("  ✅ AI-powered analysis (Kimi API)")
	fmt.Println("  ✅ CAMP-I multi-factor scoring")
	fmt.Println("  ✅ Concurrent pipeline processing")
	fmt.Println("  ✅ Multiple signal sinks (Console, JSON, Webhook)")
	fmt.Println("  ✅ Intelligent filtering system")
	fmt.Println("  ✅ Risk control (6-level validation)")
	fmt.Println("  ✅ Cost optimization (70% savings)")
}
