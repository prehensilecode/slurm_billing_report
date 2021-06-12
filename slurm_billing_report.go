package main

import (
    "fmt"
    "strings"
    "strconv"
    "flag"
    "os"
    "os/exec"
    "time"
    "runtime"
    "golang.org/x/text/language"
    "golang.org/x/text/message"
)

func execute(account string, year int, month int) {
    // USAGE REPORT FOR rosenmriprj ON CLUSTER picotte - March 2021
    cluster := "picotte"
    options := fmt.Sprintf("-n -P cluster AccountUtilizationByUser Account=%s Tree Start=%d-%02d-01 End=%d-%02d-%02d -T billing",
       account, year, month, year, month, 31)
    cmd_options := strings.Split(options, " ")
    out, err := exec.Command("sreport", cmd_options...).Output()

    if err != nil {
        fmt.Printf("%s", err)
        panic(err)
    }

    months := [12]string{"January", "February", "March", "April", "May",
        "June", "July", "August", "September", "October",
        "November", "December"}

    fmt.Printf("USAGE REPORT FOR %s ON CLUSTER %s - %s %d\n", account, cluster, months[month-1], year)

    outstr := strings.Split(string(out[:]), "\n")
    tre, err := strconv.ParseFloat(strings.Split(outstr[0], "|")[5], 64)
    su := tre / 60.

    rate := 0.0123
    charge := su * rate
    p := message.NewPrinter(language.English)
    charge_str := p.Sprintf("%.2f", charge)
    fmt.Printf("%16s %20s %18s\n", "Account", "Usage (SU)", "Charge")
    fmt.Printf("%16s %20.6e        $ %9s\n", account, su, charge_str)
}

func main() {
    debugFlag := flag.Bool("debug", false, "Debugging")
    accountFlag := flag.String("account", "", "Account/Project name")
    whenFlag := flag.String("when", "", "Period in YYYY-MM")

    flag.Parse()

    currentTime := time.Now()
    year := int(currentTime.Year())
    month := int(currentTime.Month())

    if len(*whenFlag) > 0 {
        ym := strings.Split(*whenFlag, "-")
        y := ym[0]
        m := ym[1]
        err := error(nil)
        year, err = strconv.Atoi(y)
        if err != nil {
            panic(err)
        }
        month, err = strconv.Atoi(m)
        if err != nil {
            panic(err)
        }
    }

    if *debugFlag {
        fmt.Printf("DEBUG: year-month YYYY-MM %4d-%02d\n", year, month)
    }

    if len(*accountFlag) == 0 { 
        fmt.Println("ERROR: slurm_billing_report: Must provide account name")
        os.Exit(3)
    }

    if runtime.GOOS == "windows" {
        fmt.Println("Cannot execute this on Windows")
        os.Exit(1)
    } else {
        execute(*accountFlag, year, month)
    }
}
