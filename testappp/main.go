package main

import (
	"fmt"

	"github.com/cristalhq/acmd"
)

func main() {
	info := acmd.GetBuildInfo()
	fmt.Printf("ver: %+v\n", info.Revision)
	fmt.Printf("ver: %+v\n", info.LastCommit)
	fmt.Printf("ver: %+v\n", info.IsDirtyBuild)
}
