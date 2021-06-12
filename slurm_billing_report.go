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

func execute(account string, year int, month int, debug bool) {
    // To calculate last day of a month, use fact that time.Date 
    // accepts values outside of usual ranges. So, "March 0" 
    // is the last day of February.

    end_date := time.Now()
    if month == 12 {
        end_date = time.Date(year+1, time.Month(1), 0, 0, 0, 0, 0, time.UTC)
    } else {
        end_date = time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.UTC)
    }


    options := fmt.Sprintf("-n -P cluster AccountUtilizationByUser Account=%s Tree Start=%d-%02d-01 End=%d-%02d-%02d -T billing",
       account, year, month, year, month, int(end_date.Day()))

    if debug {
        fmt.Printf("DEBUG: options = %s\n\n", options)
    }

    cmd_options := strings.Split(options, " ")
    out, err := exec.Command("sreport", cmd_options...).Output()

    if err != nil {
        fmt.Printf("%s", err)
        panic(err)
    }

    if len(out) == 0 {
        fmt.Printf("INFO: slurm_billing_report: no data for project '%s' for requested period %d-%02d\n", account, year, month)
        os.Exit(0)
    }

    rate := 0.0123
    cluster := "picotte"
    fmt.Printf("USAGE REPORT FOR %s ON CLUSTER %s - %s %d\n", account, cluster, time.Month(month), year)
    fmt.Printf("Rate = $ %.4f per SU\n\n", rate)

    outstr := strings.Split(string(out[:]), "\n")
    tre, err := strconv.ParseFloat(strings.Split(outstr[0], "|")[5], 64)
    su := tre / 60.

    charge := su * rate
    p := message.NewPrinter(language.English)
    charge_str := p.Sprintf("%.2f", charge)
    fmt.Printf("Compute usage: %8.6e SU\n", su)
    fmt.Printf("Charge: $ %9s\n", charge_str)
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

        if month > 12 {
            fmt.Printf("ERROR: slurm_billing_report: month must be <= 12 (%02d given)", month)
            os.Exit(5)
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
        execute(*accountFlag, year, month, *debugFlag)
    }
}
