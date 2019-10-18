package main

import (
	"fmt"
	"strings"
	"time"
)

var (
	Version   = ""
	GitRev    = ""
	BuildTime = ""
	Mode      = ""
)

func VersionString() string {
	if len(BuildTime) == 0 {
		BuildTime = time.Now().Format(time.RFC3339)
	}
	var b strings.Builder
	t := BuildTime
	if parse, err := time.Parse(time.RFC3339, BuildTime); err == nil {
		t = parse.Local().Format("2006-01-02 15:04:05-0700")
	}

	b.WriteString(fmt.Sprintf("%-15s%s\n", "build time:", t))
	b.WriteString(fmt.Sprintf("%-15s%s\n", "version:", Version))
	b.WriteString(fmt.Sprintf("%-15s%s\n", "hash:", GitRev))
	b.WriteString(fmt.Sprintf("%-15s%s\n", "mode:", Mode))
	return b.String()
}
