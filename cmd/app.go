package cmd

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

// Handlers contains injected command actions so CLI wiring can live outside main.
type Handlers struct {
	RunReviewSimple       cli.ActionFunc
	RunReviewDebug        cli.ActionFunc
	RunHooksInstall       cli.ActionFunc
	RunHooksUninstall     cli.ActionFunc
	RunHooksEnable        cli.ActionFunc
	RunHooksDisable       cli.ActionFunc
	RunHooksStatus        cli.ActionFunc
	RunSelfUpdate         cli.ActionFunc
	RunReviewCleanup      cli.ActionFunc
	RunAttestationTrailer cli.ActionFunc
	RunSetup              cli.ActionFunc
	RunUI                 cli.ActionFunc
}

// BuildApp constructs the full CLI app with all command wiring.
func BuildApp(version, buildTime, gitCommit string, baseFlags, debugFlags []cli.Flag, h Handlers) *cli.App {
	return &cli.App{
		Name:    "lrc",
		Usage:   "LiveReview CLI - submit local diffs for AI review",
		Version: version,
		Flags:   baseFlags,
		Commands: []*cli.Command{
			{
				Name:    "review",
				Aliases: []string{"r"},
				Usage:   "Run a review with sensible defaults",
				Flags:   baseFlags,
				Action:  h.RunReviewSimple,
			},
			{
				Name:   "review-debug",
				Usage:  "Run a review with advanced debug options",
				Flags:  append(baseFlags, debugFlags...),
				Action: h.RunReviewDebug,
			},
			{
				Name:  "hooks",
				Usage: "Manage LiveReview Git hook integration (global dispatcher)",
				Subcommands: []*cli.Command{
					{
						Name:  "install",
						Usage: "Install global LiveReview hook dispatchers (uses core.hooksPath)",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "path",
								Usage: "custom hooksPath (defaults to core.hooksPath or ~/.git-hooks)",
							},
							&cli.BoolFlag{
								Name:  "local",
								Usage: "install into the current repo hooks path (respects core.hooksPath)",
							},
						},
						Action: h.RunHooksInstall,
					},
					{
						Name:  "uninstall",
						Usage: "Remove LiveReview hook dispatchers and managed scripts",
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:  "local",
								Usage: "uninstall from the current repo hooks path",
							},
							&cli.StringFlag{
								Name:  "path",
								Usage: "target a specific hooksPath directory for uninstall",
							},
						},
						Action: h.RunHooksUninstall,
					},
					{
						Name:   "enable",
						Usage:  "Enable LiveReview hooks for the current repository",
						Action: h.RunHooksEnable,
					},
					{
						Name:   "disable",
						Usage:  "Disable LiveReview hooks for the current repository",
						Action: h.RunHooksDisable,
					},
					{
						Name:   "status",
						Usage:  "Show LiveReview hook status for the current repository",
						Action: h.RunHooksStatus,
					},
				},
			},
			{
				Name:   "install-hooks",
				Usage:  "Install LiveReview hooks (deprecated; use 'lrc hooks install')",
				Hidden: true,
				Action: h.RunHooksInstall,
			},
			{
				Name:   "uninstall-hooks",
				Usage:  "Uninstall LiveReview hooks (deprecated; use 'lrc hooks uninstall')",
				Hidden: true,
				Action: h.RunHooksUninstall,
			},
			{
				Name:  "version",
				Usage: "Show version information",
				Action: func(c *cli.Context) error {
					fmt.Printf("lrc version %s\n", version)
					fmt.Printf("  Build time: %s\n", buildTime)
					fmt.Printf("  Git commit: %s\n", gitCommit)
					return nil
				},
			},
			{
				Name:    "self-update",
				Aliases: []string{"update"},
				Usage:   "Update lrc to the latest version",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "check",
						Usage: "Only check for updates without installing",
					},
					&cli.BoolFlag{
						Name:  "force",
						Usage: "Force recovery by terminating another active lrc self-update process, then continue update",
					},
				},
				Action: h.RunSelfUpdate,
			},
			{
				Name:   "review-cleanup",
				Usage:  "Clean up review session history for the current branch (called by post-commit hook)",
				Hidden: true,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "verbose",
						Usage: "enable verbose output",
					},
				},
				Action: h.RunReviewCleanup,
			},
			{
				Name:   "attestation-trailer",
				Usage:  "Output the commit trailer for the current attestation (called by commit-msg hook)",
				Hidden: true,
				Action: h.RunAttestationTrailer,
			},
			{
				Name:   "setup",
				Usage:  "Guided onboarding — authenticate with Hexmos and configure LiveReview + AI",
				Action: h.RunSetup,
			},
			{
				Name:   "ui",
				Usage:  "Open local web UI to manage your git-lrc",
				Action: h.RunUI,
			},
		},
		Action: h.RunReviewSimple,
	}
}
