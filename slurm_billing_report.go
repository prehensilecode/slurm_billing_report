package main

import (
    "fmt"
    "strings"
    "strconv"
    "os"
    "os/exec"
    "os/user"
    "time"
    "runtime"
    "golang.org/x/text/language"
    "golang.org/x/text/message"
    "github.com/jessevdk/go-flags"
)

func execute(cluster string, account string, year int, month int, billing bool) {
    // To calculate last day of a month, use fact that time.Date 
    // accepts values outside of usual ranges. So, "March 0" 
    // is the last day of February.

    // timezone does not matter
    start_date := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
    // add 1 month to get end of period
    end_date := start_date.AddDate(0, 1, 0)

    options := fmt.Sprintf("-n -P cluster AccountUtilizationByUser Account=%s Tree Start=%d-%02d-01 End=%d-%02d-01 -T billing -t hours",
       account, int(start_date.Year()), int(start_date.Month()), int(end_date.Year()), int(end_date.Month()))

    cmd_options := strings.Split(options, " ")
    out, err := exec.Command("sreport", cmd_options...).Output()

    if err != nil {
        panic(err)
    }

    if len(out) == 0 {
        fmt.Printf("INFO: slurm_billing_report: no data for project '%s' for requested period %d-%02d\n", account, year, month)
        os.Exit(0)
    }

    pi_gecos, _ := user.Lookup(account)
    pi_name := strings.TrimSpace(strings.Split(pi_gecos.Name, "<")[0])

    outstr := strings.Split(string(out[:]), "\n")
    tre, _ := strconv.ParseFloat(strings.Split(outstr[0], "|")[5], 64)
    su := tre

    p := message.NewPrinter(language.English)

    rate := 0.
    if billing {
        rate := 0.0123
        fmt.Printf("USAGE REPORT FOR PI %s ON CLUSTER %s - %s %d\n", pi_name, cluster, time.Month(month), year)
        fmt.Printf("Rate = $ %.4f per SU\n\n", rate)
        charge := su * rate
        charge_str := p.Sprintf("%.2f", charge)
        fmt.Printf("Compute usage: %8.6e core-hours\n", su)
        fmt.Printf("Charge: $ %9s\n", charge_str)
    } else {
        fmt.Printf("USAGE REPORT FOR PI %s ON CLUSTER %s - %s %d\n", pi_name, cluster, time.Month(month), year)
        fmt.Printf("Compute usage: %8.6e core-hours\n", su)
    }


    fmt.Println("")
    fmt.Println("")

    if billing {
        fmt.Println("    Per-user usage and charge")
        fmt.Printf("%32s %12s     %12s      %9s\n", "Name", "User ID", "Usage (core-hours)", "Charge")
        for i, s := range(outstr) {
            if i > 0 && len(s) > 0 {
                line := strings.Split(s, "|")
                name := strings.TrimSpace(strings.Split(line[3], "<")[0])
                login := line[2]
                tre, _ := strconv.ParseFloat(line[5], 65)
                su := tre
                charge_str := p.Sprintf("%.2f", su * rate)
                fmt.Printf("%32s %12s     %8.6e    $ %9s\n", name, login, su, charge_str)
            }
        }
    } else {
        fmt.Println("    Per-user usage")
        fmt.Printf("%32s %12s     %12s\n", "Name", "User ID", "Usage (core-hours)")
        for i, s := range(outstr) {
            if i > 0 && len(s) > 0 {
                line := strings.Split(s, "|")
                name := strings.TrimSpace(strings.Split(line[3], "<")[0])
                login := line[2]
                tre, _ := strconv.ParseFloat(line[5], 65)
                su := tre
                fmt.Printf("%32s %12s     %8.6e\n", name, login, su)
            }
        }
    }
}

func main() {
    var opts struct {
        Account string `short:"a" long:"account" required:"true" description:"Account/Project for which to generate report (something like 'xxxxxPrj')"`
        When string `short:"w" long:"when" description:"Period for reporting in format YYYY-MM."`
        Billing bool `short:"b" long:"billing" description:"Show billing cost."`
        Cluster string `short:"c" long:"cluster" required:"true" description:"Cluster for which to generate report"`
    }


    if _, err:= flags.Parse(&opts); err != nil {
        switch flagsErr := err.(type) {
            case flags.ErrorType:
                if flagsErr == flags.ErrHelp {
                    os.Exit(0)
                }

                os.Exit(1)
            default:
                os.Exit(1)
        }
    }

    year := 0
    month := 0
    if len(opts.When) > 0 {
        ym := strings.Split(opts.When, "-")
        year, _ = strconv.Atoi(ym[0])
        month, _ = strconv.Atoi(ym[1])
        if month > 12 {
            fmt.Printf("ERROR: slurm_billing_report: month must be <= 12 (%02d given)", month)
            os.Exit(5)
        }
    } else {
        currentTime := time.Now()
        year = int(currentTime.Year())
        month = int(currentTime.Month())
    }

    if len(opts.Account) == 0 {
        fmt.Println("ERROR: slurm_billing_report: Must provide account name")
        os.Exit(3)
    }

    if len(opts.Cluster) == 0 {
        fmt.Println("ERROR: slurm_billing_report: Must provide cluster name")
        os.Exit(3)
    }

    if runtime.GOOS == "windows" {
        fmt.Println("ERROR: slurm_billing_report: Cannot run on Windows")
        os.Exit(1)
    } else {
        execute(opts.Cluster, opts.Account, year, month, opts.Billing)
    }
}
