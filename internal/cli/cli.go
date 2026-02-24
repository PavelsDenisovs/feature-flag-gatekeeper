package cli

import (
	"flag"
	"log"
)

type FlagVars struct {
}

const (
	ExitOK      = 0 // success
	ExitRuntime = 1 // runtime/internal error
	ExitUsage   = 2 // invalid CLI usage / arguments / config
)

type (
	intFlags    map[string]*int
	stringFlags map[string]*string
	boolFlags   map[string]*bool
)

// Run executes CLI and returns process exit code. It expects arguments without program name
func Run(args []string) (exitCode int) {
	if len(args) == 0 {
		log.Print("Missing command")
		return ExitUsage
	}

	intFlags, stringFlags, boolFlags := parseFlags()
	cr := newCommandRunner(args, intFlags, stringFlags, boolFlags)

	cmd := args[0]

	err := cr.runCommand(cmd)
	if err != nil {
		log.Printf("Failed to run command %s: %v", cmd, err)
		return ExitUsage
	}

	return ExitOK
}

func parseFlags() (intFlags, stringFlags, boolFlags) {
	var (
		intFlags    = make(intFlags)
		stringFlags = make(stringFlags)
		boolFlags   = make(boolFlags)
	)

	// Int flags
	intFlags["port"] = flag.Int("port", 8080, "port for the app")
	intFlags["db_port"] = flag.Int("db-port", 5433, "port for the demo database")

	// String flags
	stringFlags["config"] = flag.String("config", "", "path to the config file")

	flag.Parse()
	return intFlags, stringFlags, boolFlags
}