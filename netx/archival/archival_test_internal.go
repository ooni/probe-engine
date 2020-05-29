package archival

type DNSQueryType = dnsQueryType

func (qtype dnsQueryType) IPOfType(addr string) bool {
	return qtype.ipoftype(addr)
}
