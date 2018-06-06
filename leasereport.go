package main

// leasereport.go creates a CSV report from an existing leases file
import (
	"dhcp/leases"
	"fmt"
	"os"
)

func main() {
	// FIXME: Do more, better.

	lf, _ := os.Open("leases-snapshot.txt")
	l := leases.NewLeases(lf)

	l.Parse()
	fmt.Println("IP,MAC,Circuit ID,Binding State,Starts,Ends")
	for _, e := range l.Leases() {

		fmt.Printf("%v,%v,%v,%v,%v,%v\n",
			e.Address(),
			e.Mac(),
			e.CircuitId(),
			e.BindingState(),
			e.Starts(),
			e.Ends())
	}

}
