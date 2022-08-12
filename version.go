package main

import (
	"fmt"
	"net/http"
	"runtime/debug"
)

var version = ""
var builddate = ""

func banner(w http.ResponseWriter) {
	if version == "" {
		if build, ok := debug.ReadBuildInfo(); ok {
			version = build.Main.Version
		}
	}
	fmt.Fprintf(w, "    █████                                                       ███   ███                                        \n")
	fmt.Fprintf(w, "    ░░███                                                       ░░░   ░░░                                         \n")
	fmt.Fprintf(w, "     ░███ █████  ██████    ██████  █████████████    ██████      █████ ████   █████   ██████  ████████  █████ █████\n")
	fmt.Fprintf(w, "     ░███░░███  ░░░░░███  ███░░███░░███░░███░░███  ███░░███    ░░███ ░░███  ███░░   ███░░███░░███░░███░░███ ░░███ \n")
	fmt.Fprintf(w, "     ░██████░    ███████ ░███ ░███ ░███ ░███ ░███ ░███ ░███     ░███  ░███ ░░█████ ░███████  ░███ ░░░  ░███  ░███ \n")
	fmt.Fprintf(w, "     ░███░░███  ███░░███ ░███ ░███ ░███ ░███ ░███ ░███ ░███     ░███  ░███  ░░░░███░███░░░   ░███      ░░███ ███  \n")
	fmt.Fprintf(w, "     ████ █████░░████████░░██████  █████░███ █████░░██████      ░███  █████ ██████ ░░██████  █████      ░░█████   \n")
	fmt.Fprintf(w, "    ░░░░ ░░░░░  ░░░░░░░░  ░░░░░░  ░░░░░ ░░░ ░░░░░  ░░░░░░       ░███ ░░░░░ ░░░░░░   ░░░░░░  ░░░░░        ░░░░░    \n")
	fmt.Fprintf(w, "                                                            ███ ░███                                              \n")
	fmt.Fprintf(w, "                                                           ░░██████                                               %s\n", version)
	fmt.Fprintf(w, "                                                            ░░░░░░                                                \n")
}
