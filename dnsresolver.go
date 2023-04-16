package main

import (
	"fmt"
	"net"
	"sync"

	"github.com/miekg/dns"
)

var recordTypes = []string{
	"A",
	"AAAA",
	"CNAME",
	"MX",
	"NS",
	"PTR",
	"SOA",
	"SRV",
	"TXT",
}

func FetchAllRecords(domain string) ([]string, error) {
	var records []string
	var wg sync.WaitGroup
	ch := make(chan []string)

	for _, recordType := range recordTypes {
		wg.Add(1)
		go func(r string) {
			defer wg.Done()
			var recs []string
			var err error

			switch r {
			case "A":
				recs, err = lookupA(domain)
			case "AAAA":
				recs, err = lookupAAAA(domain)
			case "CNAME":
				recs, err = lookupCNAME(domain)
			case "MX":
				recs, err = lookupMX(domain)
			case "NS":
				recs, err = lookupNS(domain)
			case "PTR":
				recs, err = lookupPTR(domain)
			case "SOA":
				recs, err = lookupSOA(domain)
			case "SRV":
				recs, err = lookupSRV(domain)
			case "TXT":
				recs, err = lookupTXT(domain)
			}

			if err == nil && len(recs) > 0 {
				ch <- recs
			}
		}(recordType)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for recs := range ch {
		records = append(records, recs...)
	}

	return records, nil
}

func lookupA(domain string) ([]string, error) {
	ips, err := net.LookupIP(domain)
	if err != nil {
		return nil, err
	}
	var records []string
	for _, ip := range ips {
		if ip.To4() != nil {
			records = append(records, fmt.Sprintf("A %s", ip.String()))
		}
	}
	return records, nil
}

func lookupAAAA(domain string) ([]string, error) {
	ips, err := net.LookupIP(domain)
	if err != nil {
		return nil, err
	}
	var records []string
	for _, ip := range ips {
		if ip.To4() == nil {
			records = append(records, fmt.Sprintf("AAAA %s", ip.String()))
		}
	}
	return records, nil
}

func lookupCNAME(domain string) ([]string, error) {
	cname, err := net.LookupCNAME(domain)
	if err != nil {
		return nil, err
	}
	return []string{fmt.Sprintf("CNAME %s", cname)}, nil
}

func lookupMX(domain string) ([]string, error) {
	mxRecords, err := net.LookupMX(domain)
	if err != nil {
		return nil, err
	}
	var records []string
	for _, mx := range mxRecords {
		records = append(records, fmt.Sprintf("MX %s %d", mx.Host, mx.Pref))
	}
	return records, nil
}

func lookupNS(domain string) ([]string, error) {
	nsRecords, err := net.LookupNS(domain)
	if err != nil {
		return nil, err
	}
	var records []string
	for _, ns := range nsRecords {
		records = append(records, fmt.Sprintf("NS %s", ns.Host))
	}
	return records, nil
}

func lookupPTR(domain string) ([]string, error) {
	ptrRecords, err := net.LookupAddr(domain)
	if err != nil {
		return nil, err
	}
	var records []string
	for _, ptr := range ptrRecords {
		records = append(records, fmt.Sprintf("PTR %s", ptr))
	}
	return records, nil
}

func lookupSOA(domain string) ([]string, error) {
	client := dns.Client{}
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeSOA)

	res, _, err := client.Exchange(msg, "8.8.8.8:53")
	if err != nil {
		return nil, err
	}

	var records []string
	for _, answer := range res.Answer {
		if soa, ok := answer.(*dns.SOA); ok {
			record := fmt.Sprintf("SOA %s %s %d %d %d %d %d", soa.Ns, soa.Mbox, soa.Serial, soa.Refresh, soa.Retry, soa.Expire, soa.Minttl)
			records = append(records, record)
		}
	}
	return records, nil
}

func lookupSRV(domain string) ([]string, error) {
	// Replace _service._proto with the desired service and protocol.
	// Example: _sip._tcp
	_, srvRecords, err := net.LookupSRV("_service._proto", "tcp", domain)
	if err != nil {
		return nil, err
	}
	var records []string
	for _, srv := range srvRecords {
		records = append(records, fmt.Sprintf("SRV %s %d %d %d", srv.Target, srv.Port, srv.Priority, srv.Weight))
	}
	return records, nil
}

func lookupTXT(domain string) ([]string, error) {
	txtRecords, err := net.LookupTXT(domain)
	if err != nil {
		return nil, err
	}
	var records []string
	for _, txt := range txtRecords {
		records = append(records, fmt.Sprintf("TXT \"%s\"", txt))
	}
	return records, nil
}
