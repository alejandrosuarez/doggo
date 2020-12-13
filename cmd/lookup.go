package main

import (
	"strings"

	"github.com/miekg/dns"
	"github.com/mr-karan/doggo/pkg/resolvers"
	"github.com/sirupsen/logrus"
)

// Lookup sends the DNS queries to the server.
func (hub *Hub) Lookup() error {
	err := hub.prepareQuestions()
	if err != nil {
		return err
	}
	responses, err := hub.Resolver.Lookup(hub.Questions)
	if err != nil {
		return err
	}
	hub.Output(responses)
	return nil
}

// prepareQuestions iterates on list of domain names
// and prepare a list of questions
// sent to the server with all possible combinations.
func (hub *Hub) prepareQuestions() error {
	var (
		question dns.Question
	)
	for _, name := range hub.QueryFlags.QNames {
		var (
			domains []string
			ndots   int
		)
		ndots = 1

		// If `search` flag is specified then fetch the search list
		// from `resolv.conf` and set the
		if hub.QueryFlags.UseSearchList {
			list, n, err := fetchDomainList(name, false, hub.QueryFlags.Ndots)
			if err != nil {
				return err
			}
			domains = list
			ndots = n
		} else {
			domains = []string{dns.Fqdn(name)}
		}
		for _, d := range domains {
			hub.Logger.WithFields(logrus.Fields{
				"domain": d,
				"ndots":  ndots,
			}).Debug("Attmepting to resolve")
			question.Name = d
			// iterate on a list of query types.
			for _, q := range hub.QueryFlags.QTypes {
				question.Qtype = dns.StringToType[strings.ToUpper(q)]
				// iterate on a list of query classes.
				for _, c := range hub.QueryFlags.QClasses {
					question.Qclass = dns.StringToClass[strings.ToUpper(c)]
					// append a new question for each possible pair.
					hub.Questions = append(hub.Questions, question)
				}
			}
		}
	}
	return nil
}

func fetchDomainList(d string, isNdotsSet bool, ndots int) ([]string, int, error) {
	cfg, err := dns.ClientConfigFromFile(resolvers.DefaultResolvConfPath)
	if err != nil {
		return nil, 0, err
	}
	// if user specified a custom ndots parameter, override it
	if isNdotsSet {
		cfg.Ndots = ndots
	}
	return cfg.NameList(d), cfg.Ndots, nil
}