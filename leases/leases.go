// Package leases performs simple parsing of ISC DHCP server's
// dhcpd.leases files
package leases

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"time"
)

const (
	ISCDHCPD = "2006/01/02 15:04:05"
)

type Lease struct {
	address   string    // The IP address assigned
	starts    time.Time // Lease start time
	ends      time.Time // Lease end time
	binding   string    // Lease binding state
	mac       string    // Hardware/MAC address
	circuitId string    // DHCP Option 82 - Circuit ID
	e         []error   // Parse errors
}

type LeaseFile struct {
	scanner *bufio.Scanner //The reader provides the raw source of the dhcpd.leases file
	leases  map[string]*Lease
}

// NewLeases creates a new *LeaseFile to store parsed dhcpd.leases entries.
// A new bufio.Scanner is created and configured to use the ScanLeases split
// function.
func NewLeases(r io.Reader) *LeaseFile {
	s := bufio.NewScanner(r)
	s.Split(ScanLeases)
	l := new(LeaseFile)
	l.scanner = s
	l.leases = make(map[string]*Lease)
	return l
}

func (l *Lease) setAddress(s string) {
	l.address = s
}

func (l Lease) Address() string {
	return l.address
}

// setMac extracts a lease's associated hardware MAC address.
func (l *Lease) setMac(s string) {
	l.mac = s[20 : len(s)-1]
}

func (l Lease) Mac() string {
	return l.mac
}

// setCircuitId extracts the circuit ID from a DHCP leases option 82 entry
func (l *Lease) setCircuitId(s string) {
	l.circuitId = s[27 : len(s)-2]
}

func (l Lease) CircuitId() string {
	return l.circuitId
}

// setBindingState extracts a lease's binding state.
func (l *Lease) setBindingState(s string) {
	l.binding = s[16 : len(s)-1]
}

func (l Lease) BindingState() string {
	return l.binding
}

// setStarts extracts a lease's start date/time.
// Extracted time string is converted to a time.Time type.
func (l *Lease) setStarts(s string) {
	t, err := LeaseTime(s[11 : len(s)-1])
	if err != nil {
		l.e = append(l.e, err)
	}
	l.starts = t
}

// Ends returns a lease's start date/time as time.Time.
func (l Lease) Starts() time.Time {
	return l.starts
}

// setEnds extracts a lease's end date/time.
// Extracted time string is converted to a time.Time type.
func (l *Lease) setEnds(s string) {
	t, err := LeaseTime(s[9 : len(s)-1])
	if err != nil {
		l.e = append(l.e, err)
	}
	l.ends = t
}

// Ends returns a lease's end date/time as time.Time.
func (l Lease) Ends() time.Time {
	return l.ends
}

func (l Lease) Active() bool {
	return time.Now().Before(l.ends)
}

func (l Lease) Error() []error {
	return l.e
}

func (lf *LeaseFile) AddLease(lease *Lease) {
	lf.leases[lease.Address()] = lease
}

func (lf LeaseFile) Leases() map[string]*Lease {
	return lf.leases
}

func (lf *LeaseFile) Parse() {

	for lf.scanner.Scan() {
		entry := strings.NewReader(lf.scanner.Text())

		s := bufio.NewScanner(entry)

		for s.Scan() {
			lease := new(Lease)
			if strings.Contains(s.Text(), "lease") {

				lease.setAddress(strings.Fields(s.Text())[1])
				for s.Scan() {
					if strings.Contains(s.Text(), "ends") {
						lease.setEnds(s.Text())
					}

					if strings.Contains(s.Text(), "starts") {
						lease.setStarts(s.Text())
					}

					if strings.Contains(s.Text(), "hardware") {
						lease.setMac(s.Text())
					}

					if strings.Contains(s.Text(), "option agent.circuit-id") {
						lease.setCircuitId(s.Text())
					}

					if strings.Contains(s.Text(), "  binding state") {
						lease.setBindingState(s.Text())
					}

				}
			}
			if len(lease.Error()) > 0 {
				continue
			}

			if len(lease.Address()) > 0 {
				lf.AddLease(lease)
			}

		}

	}

}

// LeaseTime parses the dhcpd.leases date format into time.Time
// BUG(dkm): Currently assumes location is America/New_York.
// BUG(dkm): I haven't yet confirmed whether lease times in dhcpd.leases are UTC.
func LeaseTime(d string) (time.Time, error) {
	loc, _ := time.LoadLocation("America/New_York")
	return time.ParseInLocation(ISCDHCPD, d, loc)
}

// ScanLeases is a custom split function for bufio.Scanner.
// The lease file splits on the closing bracket and returns
// a token containing the full text of a leases entry with the trailing
// curly brace removed.
func ScanLeases(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '}'); i >= 0 {
		return i + 1, data[0:i], nil
	}

	// Request more data.
	return 0, nil, nil
}
