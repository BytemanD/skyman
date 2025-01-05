package guest

import (
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/BytemanD/go-console/console"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

const Mb = 1000
const MB = Mb * 8
const Gb = Mb * 1000
const GB = Gb * 8
const Tb = Gb * 1000
const TB = Tb * 8

type Bandwidth struct {
	Value float64
	Unit  string
}
type PPS struct {
	Lost  int
	Value int
}

func (pps PPS) GetLostPercent() float64 {
	return float64(pps.Lost) * 100 / float64(pps.Value)
}

func (pps *PPS) Sum(other PPS) {
	pps.Lost += other.Lost
	pps.Value += other.Value
}

func (bandwidth Bandwidth) ToKbits() float64 {
	switch bandwidth.Unit {
	case "Mbits/sec":
		return bandwidth.Value * Mb
	case "Gbits/sec":
		return bandwidth.Value * Gb
	case "Tbits/sec":
		return bandwidth.Value * Gb
	case "KBytes/sec":
		return bandwidth.Value * 8
	case "MBytes/sec":
		return bandwidth.Value * MB
	case "GBytes/sec":
		return bandwidth.Value * GB
	case "TBytes/sec":
		return bandwidth.Value * GB
	default:
		return bandwidth.Value
	}
}
func (bandwidth Bandwidth) String() string {
	return fmt.Sprintf("%.2f %s", bandwidth.Value, bandwidth.Unit)
}
func (bandwidth Bandwidth) HumanBandwidth() string {
	return ParseBandwidth(bandwidth.ToKbits())
}
func ParseBandwidth(valueKb float64) string {
	switch {
	case valueKb >= Gb:
		return fmt.Sprintf("%.2f Gbits/sec", valueKb/Gb)
	case valueKb >= Mb:
		return fmt.Sprintf("%.2f Mbits/sec", valueKb/Mb)
	default:
		return fmt.Sprintf("%.2f Kbits/sec", valueKb)
	}
}

type IperfReport struct {
	Source string
	Dest   string
	Data   string
}

func (r IperfReport) GetBpsSum() (*Bandwidth, *Bandwidth) {
	senderReg := regexp.MustCompile(" +([0-9.]+) ([a-zA-Z]+/sec) .* +sender")
	receiverReg := regexp.MustCompile(" +([0-9.]+) ([a-zA-Z]+/sec) .* +receiver")

	matchedSenders := senderReg.FindAllStringSubmatch(r.Data, -1)
	matchedReceivers := receiverReg.FindAllStringSubmatch(r.Data, -1)
	if len(matchedSenders) == 0 || len(matchedReceivers) == 0 {
		return nil, nil
	}
	// NOTE: Only get the last matched item
	numberSend, _ := strconv.ParseFloat(
		matchedSenders[len(matchedSenders)-1][1], 64)
	sumSend := Bandwidth{
		Value: numberSend,
		Unit:  matchedSenders[len(matchedSenders)-1][2]}
	numberReceive, _ := strconv.ParseFloat(
		matchedReceivers[len(matchedReceivers)-1][1], 64)
	sumReceive := Bandwidth{
		Value: numberReceive,
		Unit:  matchedReceivers[len(matchedReceivers)-1][2]}

	return &sumSend, &sumReceive
}
func (r IperfReport) GetPpsSum() (*PPS, *PPS) {
	senderReg := regexp.MustCompile(" +([0-9]+)/([0-9]+) \\([0-9]+%\\) +sender")
	receiverReg := regexp.MustCompile(" +([0-9]+)/([0-9]+) \\([0-9]+%\\) +receiver")

	matchedSenders := senderReg.FindAllStringSubmatch(r.Data, -1)
	matchedReceivers := receiverReg.FindAllStringSubmatch(r.Data, -1)
	if len(matchedSenders) == 0 || len(matchedReceivers) == 0 {
		console.Warn("sender or receiver not found")
		return nil, nil
	}
	// NOTE: Only get the last matched item
	sumSendLost, _ := strconv.Atoi(matchedSenders[len(matchedSenders)-1][1])
	sumSend, _ := strconv.Atoi(matchedSenders[len(matchedSenders)-1][2])

	sumReceiveLost, _ := strconv.Atoi(matchedReceivers[len(matchedReceivers)-1][1])
	sumReceive, _ := strconv.Atoi(matchedReceivers[len(matchedReceivers)-1][2])

	return &PPS{Lost: sumSendLost, Value: sumSend},
		&PPS{Lost: sumReceiveLost, Value: sumReceive}
}

func NewIperfReport(source string, dest string, data string) *IperfReport {
	return &IperfReport{Source: source, Dest: dest, Data: data}
}

type IperfReports struct {
	Reports      []IperfReport
	SendTotal    Bandwidth
	ReceiveTotal Bandwidth
}

func NewIperfReports() *IperfReports {
	return &IperfReports{}
}

func (reports *IperfReports) Add(source string, dest string, data string) {
	reports.Reports = append(reports.Reports,
		IperfReport{Source: source, Dest: dest, Data: data})
}
func (reports IperfReports) PrintBps() {
	tableWriter := table.NewWriter()
	tableWriter.SetOutputMirror(os.Stdout)
	tableWriter.SetAutoIndex(true)
	tableWriter.SetStyle(table.StyleLight)
	tableWriter.Style().Format.Header = text.FormatDefault
	tableWriter.Style().Format.Footer = text.FormatDefault
	tableWriter.Style().Options.SeparateRows = true
	tableWriter.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, AutoMerge: true, AlignHeader: text.AlignCenter},
		{Number: 2, AutoMerge: true, AlignHeader: text.AlignCenter},
		{Number: 3, AutoMerge: true, AlignHeader: text.AlignCenter},
	})
	rowConfigAutoMerge := table.RowConfig{AutoMerge: true}
	tableWriter.AppendHeader(
		table.Row{"Client -> Server", "Bandwidth", "Bandwidth"},
		rowConfigAutoMerge,
	)
	tableWriter.AppendHeader(
		table.Row{"", "Sender", "Receiver"},
		rowConfigAutoMerge, rowConfigAutoMerge, rowConfigAutoMerge,
	)

	console.Debug("reports %v", reports.Reports)
	for _, report := range reports.Reports {
		sumSend, sumReceive := report.GetBpsSum()
		if sumSend == nil || sumReceive == nil {
			console.Warn("%s -> %s sum sender or receiver is not found",
				report.Source, report.Dest)
			continue
		}
		reports.SendTotal.Value += sumSend.ToKbits()
		reports.ReceiveTotal.Value += sumReceive.ToKbits()
		tableWriter.AppendRow(
			table.Row{
				fmt.Sprintf("%s -> %s", report.Source, report.Dest),
				sumSend.String(), sumReceive.String(),
			},
		)
	}
	tableWriter.AppendFooter(table.Row{
		"Total", reports.SendTotal.HumanBandwidth(), reports.ReceiveTotal.HumanBandwidth(),
	})
	tableWriter.Render()
}
func (reports IperfReports) PrintPps() {
	tableWriter := table.NewWriter()
	tableWriter.SetOutputMirror(os.Stdout)
	tableWriter.SetAutoIndex(true)
	tableWriter.SetStyle(table.StyleLight)
	tableWriter.Style().Format.Header = text.FormatDefault
	tableWriter.Style().Format.Footer = text.FormatDefault
	tableWriter.Style().Options.SeparateRows = true
	tableWriter.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, AutoMerge: true, AlignHeader: text.AlignCenter},
		{Number: 2, AutoMerge: true, AlignHeader: text.AlignCenter},
		{Number: 3, AutoMerge: true, AlignHeader: text.AlignCenter},
	})
	rowConfigAutoMerge := table.RowConfig{AutoMerge: true}
	tableWriter.AppendHeader(
		table.Row{"Client -> Server", "PPS", "PPS"},
		rowConfigAutoMerge,
	)
	tableWriter.AppendHeader(
		table.Row{"", "Sender(Lost/Total)", "Receiver(Lost/Total)"},
		rowConfigAutoMerge, rowConfigAutoMerge, rowConfigAutoMerge,
	)
	var (
		sTotal PPS
		rTotal PPS
	)
	console.Debug("reports %v", reports.Reports)
	for _, report := range reports.Reports {
		sumSend, sumReceive := report.GetPpsSum()
		sTotal.Sum(*sumSend)
		rTotal.Sum(*sumReceive)

		tableWriter.AppendRow(
			table.Row{
				fmt.Sprintf("%s -> %s", report.Source, report.Dest),
				fmt.Sprintf("%d/%d", sumSend.Lost, sumSend.Value),
				fmt.Sprintf("%d/%d", sumReceive.Lost, sumReceive.Value),
			},
		)
	}
	tableWriter.AppendFooter(
		table.Row{"Total",
			fmt.Sprintf("%d/%d (%.2f)", sTotal.Lost, sTotal.Value, sTotal.GetLostPercent()),
			fmt.Sprintf("%d/%d (%.2f)", rTotal.Lost, rTotal.Value, rTotal.GetLostPercent())},
	)
	tableWriter.Render()
}
