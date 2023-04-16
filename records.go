package main

type DNSRecord struct {
	Type string
	Data string
}

func (r *DNSRecord) String() string {
	return r.Type + " " + r.Data
}
