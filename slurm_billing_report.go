package main

import (
    "fmt"
    "strings"
    "strconv"
    "os/exec"
    "runtime"
)

func execute() {
    options := fmt.Sprintf("-n -P cluster AccountUtilizationByUser Account=%s Tree Start=%d-%02d-01 End=%d-%02d-%02d -T billing",
       "rosenmriprj", 2021, 3, 2021, 3, 31)
    cmd_options := strings.Split(options, " ")
    out, err := exec.Command("sreport", cmd_options...).Output()

    if err != nil {
        fmt.Printf("%s", err)
    }

    outstr := strings.Split(string(out[:]), "\n")
    tre, err := strconv.ParseFloat(strings.Split(outstr[0], "|")[5], 64)
    su := tre / 60.

    rate := 0.0123
    charge := su * rate
    fmt.Printf("SU %.6e    Charge  $ %.2f\n", su, charge)
}

func main() {
    if runtime.GOOS == "windows" {
        fmt.Println("Cannot execute this on Windows")
    } else {
        execute()
    }
}
