package main

import (
	"flag"
	"fmt"
	"github.com/qlcchain/go-qlc/vm/exec"
	"github.com/qlcchain/go-qlc/vm/platform"
	"github.com/qlcchain/go-qlc/vm/resolver"
	"io/ioutil"
	"strconv"
	"time"
)

func main() {
	entryFunctionFlag := flag.String("entry", "app_main", "entry function name")
	pmFlag := flag.Bool("polymerase", false, "enable the Polymerase engine")
	noFloatingPointFlag := flag.Bool("no-fp", false, "disable floating point")
	flag.Parse()

	// Read WebAssembly *.wasm file.
	input, err := ioutil.ReadFile(flag.Arg(0))
	if err != nil {
		panic(err)
	}

	// Instantiate a new WebAssembly VM with a few resolved imports.
	vm, err := exec.NewVirtualMachine(input, exec.VMConfig{
		DefaultMemoryPages:   128,
		DefaultTableSize:     65536,
		DisableFloatingPoint: *noFloatingPointFlag,
	}, resolver.NewResolver(), nil)

	if err != nil {
		panic(err)
	}

	if *pmFlag {
		compileStartTime := time.Now()
		fmt.Println("[Polymerase] Compilation started.")
		aotSvc := platform.FullAOTCompile(vm)
		if aotSvc != nil {
			compileEndTime := time.Now()
			fmt.Printf("[Polymerase] Compilation finished successfully in %+v.\n", compileEndTime.Sub(compileStartTime))
			vm.SetAOTService(aotSvc)
		} else {
			fmt.Println("[Polymerase] The current platform is not yet supported.")
		}
	}

	// Get the function ID of the entry function to be executed.
	entryID, ok := vm.GetFunctionExport(*entryFunctionFlag)
	if !ok {
		fmt.Printf("Entry function %s not found; starting from 0.\n", *entryFunctionFlag)
		entryID = 0
	}

	start := time.Now()

	// If any function prior to the entry function was declared to be
	// called by the module, run it first.
	if vm.Module.Base.Start != nil {
		startID := int(vm.Module.Base.Start.Index)
		_, err := vm.Run(startID)
		if err != nil {
			vm.PrintStackTrace()
			panic(err)
		}
	}
	var args []int64
	for _, arg := range flag.Args()[1:] {
		fmt.Println(arg)
		if ia, err := strconv.Atoi(arg); err != nil {
			panic(err)
		} else {
			args = append(args, int64(ia))
		}
	}

	// Run the WebAssembly module's entry function.
	ret, err := vm.Run(entryID, args...)
	if err != nil {
		vm.PrintStackTrace()
		panic(err)
	}
	end := time.Now()

	fmt.Printf("return value = %d, duration = %v\n", ret, end.Sub(start))
}
