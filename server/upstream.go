package server

import "github.com/miekg/dns"
import "math/rand"
import "log"
import "time"

// RandomUpstream resolves using a random NS in the set.
type RandomUpstream struct {
	upstream []string
}

// ServeDNS resolution.
func (h *RandomUpstream) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {

	client := &dns.Client{
		Net: w.RemoteAddr().Network(),
	}

	var res *dns.Msg
	var rtt time.Duration
	var err error
	
	idx := rand.Intn(len(h.upstream))
	ns := h.upstream[idx]
	ns = defaultPort(ns)

	for try:=0; try<len(h.upstream); try++ {
		for _, q := range r.Question {
			log.Printf("[info] [%v] <== %s %s %v (ns %s)\n", r.Id,
				dns.ClassToString[q.Qclass],
				dns.TypeToString[q.Qtype],
				q.Name,
				ns)
		}

		res, rtt, err = client.Exchange(r, ns)

		if err == nil {
			break;
		}

		// if all ns failed to exchange
		if try == len(h.upstream)-1 {
			log.Printf("[error] [%v] failed to exchange – %s", r.Id, err)

			msg := new(dns.Msg)
			msg.SetRcode(r, dns.RcodeServerFailure)
			w.WriteMsg(msg)
			return
		}

		// use next ns
		idx = (idx+1) % len(h.upstream)
		ns = h.upstream[idx]
		ns = defaultPort(ns)

	}

	log.Printf("[info] [%v] ==> %s:", r.Id, rtt)
	for _, a := range res.Answer {
		log.Printf("[info] [%v] ----> %s\n", r.Id, a)
	}

	err = w.WriteMsg(res)
	if err != nil {
		log.Printf("[error] [%v] failed to respond – %s", r.Id, err)
	}
}
