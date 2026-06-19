package reviewquery

import (
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"
)

// RunQuery is the default action for `lrc query`. It either saves an alias
// (--add/--name) or runs a saved alias / raw SQL and prints a table or JSON.
func RunQuery(c *cli.Context) error {
	if c.IsSet("add") {
		add := strings.TrimSpace(c.String("add"))
		name := strings.TrimSpace(c.String("name"))
		if !c.IsSet("name") || name == "" {
			return fmt.Errorf("--add requires --name")
		}
		if add == "" {
			return fmt.Errorf("--add requires non-empty SQL")
		}
		if err := AddAlias(name, add); err != nil {
			return err
		}
		fmt.Printf("Saved alias %q.\n", name)
		return nil
	}

	// Seed from flags placed BEFORE the positional arg (cli parses those).
	jsonOut := c.Bool("json")
	filter := Filter{From: c.String("from"), To: c.String("to"), Range: c.String("range")}

	// urfave/cli stops parsing flags at the first positional arg, so also scan
	// the remaining args for trailing flags (e.g. `lrc query stats --from 2024-01-01`).
	positionals, err := parseTrailingFlags(c.Args().Slice(), &jsonOut, &filter)
	if err != nil {
		return err
	}

	arg := "stats" // default alias
	if len(positionals) > 0 && strings.TrimSpace(positionals[0]) != "" {
		arg = strings.TrimSpace(positionals[0])
	}

	sqlText, found, err := ResolveAlias(arg)
	if err != nil {
		return err
	}
	if !found {
		// Not a known alias — treat the positional args as raw SQL.
		sqlText = strings.Join(positionals, " ")
	}

	res, err := Run(filter, sqlText)
	if err != nil {
		return err
	}

	if jsonOut {
		out, err := FormatJSON(res)
		if err != nil {
			return err
		}
		fmt.Println(out)
		return nil
	}
	fmt.Print(FormatTable(res))
	return nil
}

// parseTrailingFlags pulls flags out of args that cli left unparsed (anything
// after the first positional). Supports `--flag value` and `--flag=value`.
// Returns the remaining positional args; sets jsonOut/filter via pointers.
// Returns an error if a bound flag (--from/--to/--range) is the last arg with
// no value following it, rather than silently swallowing the flag name into
// the positionals (where it would end up mangling the SQL/alias lookup).
func parseTrailingFlags(args []string, jsonOut *bool, filter *Filter) ([]string, error) {
	boundFlags := []struct {
		name string
		dest *string
	}{
		{"--from", &filter.From},
		{"--to", &filter.To},
		{"--range", &filter.Range},
	}

	positionals := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--json" || a == "-j" {
			*jsonOut = true
			continue
		}

		consumed := false
		for _, bf := range boundFlags {
			if val, ok := strings.CutPrefix(a, bf.name+"="); ok {
				*bf.dest = val
				consumed = true
				break
			}
			if a == bf.name {
				if i+1 >= len(args) {
					return nil, fmt.Errorf("%s requires a value", bf.name)
				}
				*bf.dest = args[i+1]
				i++
				consumed = true
				break
			}
		}
		if consumed {
			continue
		}
		positionals = append(positionals, a)
	}
	return positionals, nil
}

// RunQueryList prints every alias and its source.
func RunQueryList(c *cli.Context) error {
	aliases, err := ListAliases()
	if err != nil {
		return err
	}
	if len(aliases) == 0 {
		fmt.Println("(no aliases)")
		return nil
	}
	for _, a := range aliases {
		fmt.Printf("%-18s [%s]\n", a.Name, a.Source)
	}
	return nil
}

// RunQueryView prints the SQL behind a named alias.
func RunQueryView(c *cli.Context) error {
	name := strings.TrimSpace(c.Args().First())
	if name == "" {
		return fmt.Errorf("usage: lrc query view <name>")
	}
	sqlText, found, err := ResolveAlias(name)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("no alias named %q", name)
	}
	fmt.Println(sqlText)
	return nil
}

// RunQueryDelete removes a user-defined alias.
func RunQueryDelete(c *cli.Context) error {
	name := strings.TrimSpace(c.Args().First())
	if name == "" {
		return fmt.Errorf("usage: lrc query delete <name>")
	}
	if err := DeleteAlias(name); err != nil {
		return err
	}
	fmt.Printf("Deleted alias %q.\n", name)
	return nil
}
